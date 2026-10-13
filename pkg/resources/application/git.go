package application

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

// GitDeploy pushes the configured repository to the Clever Cloud remote.
// deployedCommit is the commit currently deployed on the application (CommitID
// from the API, empty when unknown or never deployed): when the resolved
// target commit already matches it, the push is skipped.
// It returns the resolved target commit (empty when no deployment is
// configured, the repository is handled by Github or the resolution failed)
// and true only when a push actually triggered a deployment, so the caller
// can decide whether an explicit restart is still needed (e.g. when only
// environment variables changed).
func GitDeploy(ctx context.Context, d *Deployment, cleverRemote, deployedCommit string, diags *diag.Diagnostics) (string, bool) {
	var errs diag.Diagnostics
	var commit string
	var deployed bool

	if d == nil {
		return "", false
	}
	if d.Commit != nil && strings.HasPrefix(*d.Commit, attributes.GITHUB_COMMIT_PREFIX) {
		tflog.Warn(ctx, "repository deployment is handled by Github, skipping deployment")
		return "", false
	}

	for range 5 {
		commit, deployed, errs = gitDeploy(ctx, *d, cleverRemote, deployedCommit)
		if !errs.HasError() || !retryable(errs) {
			break
		}

		time.Sleep(3 * time.Second)
	}

	// only add last error
	diags.Append(errs...)
	return commit, deployed
}

func gitDeploy(ctx context.Context, d Deployment, cleverRemote, deployedCommit string) (string, bool, diag.Diagnostics) {
	cleverRemote = strings.Replace(cleverRemote, "git+ssh", "https", 1) // switch protocol

	repo, diags := OpenOrClone(ctx, d.Repository, WithCommit(d.Commit), WithBasicAuth(d.Username, d.Password))
	if diags.HasError() {
		return "", false, diags
	}

	targetCommit, diags := resolveTargetCommit(repo, d.Commit)
	if diags.HasError() {
		return "", false, diags
	}

	if deployedCommit != "" && targetCommit == deployedCommit {
		tflog.Info(ctx, "deployed commit is already the expected one, skipping git push", map[string]any{
			"commit": targetCommit,
		})
		return targetCommit, false, diags
	}

	var auth transport.AuthMethod
	if d.CleverGitAuth != nil {
		auth = d.CleverGitAuth
	}

	tflog.Debug(ctx, "pushing...", map[string]any{
		"remote": cleverRemote,
		"commit": targetCommit,
	})

	pushed, err := streamPush(ctx, repo.Storer, cleverRemote, auth, plumbing.NewHash(targetCommit), os.Stdout)
	var rejected *receivePackError
	if errors.As(err, &rejected) && rejected.tooLarge() {
		// pushing the same pack again gets the same answer
		diags.Append(nonRetryable{diag.NewErrorDiagnostic(
			"repository too large for the clever remote",
			fmt.Sprintf("the push of commit %s was rejected by the Clever Cloud git server: %s", targetCommit, err.Error()),
		)})
		return "", false, diags
	}
	if err != nil {
		diags.AddError("failed to push to clever remote", err.Error())
		return "", false, diags
	}
	if !pushed {
		diags.AddWarning("Git push rejected", "repository is already up-to-date")
		return targetCommit, false, diags
	}

	return targetCommit, true, diags
}

// nonRetryable marks an error that another push attempt would get again
type nonRetryable struct {
	diag.Diagnostic
}

// retryable tells whether a failed deployment is worth another attempt
func retryable(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if _, ok := d.(nonRetryable); ok {
			return false
		}
	}
	return true
}

// resolveTargetCommit resolves the commit hash to deploy: the explicit
// commit/reference when one is set, the repository HEAD otherwise.
func resolveTargetCommit(repo *git.Repository, commit *string) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if commit == nil {
		head, err := repo.Head()
		if err != nil {
			diags.AddError("failed to get current ref", err.Error())
			return "", diags
		}

		return head.Hash().String(), diags
	}

	refNameOrCommit := *commit

	// can be
	// refs/heads/[BRANCH]
	// or
	// [COMMIT_SHA] a397296e135b24e682a011e31f8e15f2fa8a5a0e

	if IsSHA1(refNameOrCommit) {
		c, err := repo.CommitObject(plumbing.NewHash(refNameOrCommit))
		if err == plumbing.ErrObjectNotFound {
			diags.AddError("requested commit not found", fmt.Sprintf("no commit '%s'", refNameOrCommit))
			return "", diags
		}
		if err != nil {
			diags.AddError("failed to look for commit", err.Error())
			return "", diags
		}

		return c.Hash.String(), diags
	}

	if !strings.HasPrefix(refNameOrCommit, "refs/") {
		refNameOrCommit = "refs/heads/" + refNameOrCommit
	}

	// We need to check if provided ref exists (several issues with main/master)
	ref, err := repo.Storer.Reference(plumbing.ReferenceName(refNameOrCommit))
	if err == plumbing.ErrReferenceNotFound {
		diags.AddError("requested reference not found", fmt.Sprintf("no reference named '%s'", refNameOrCommit))
		return "", diags
	}
	if err != nil {
		diags.AddError("failed to get reference", err.Error())
		return "", diags
	}

	return ref.Hash().String(), diags
}

func IsSHA1(s string) bool {
	h, err := hex.DecodeString(s)
	return err == nil && len(h) == sha1.Size
}

func OpenOrClone(ctx context.Context, repoUrl string, opts ...CloneOpts) (*git.Repository, diag.Diagnostics) {
	if strings.HasPrefix(repoUrl, "file://") {
		return open(repoUrl)
	}

	return clone(ctx, repoUrl, opts...)
}

// largeObjectThreshold is the size above which objects of a local repository
// are streamed from disk instead of being loaded in memory. The push reads one
// object at a time, so below it the memory only depends on the largest object.
// It is not lower because go-git rebuilds a delta-stored object very slowly
// when streaming it (about a minute for a 36 MiB binary).
const largeObjectThreshold = 64 << 20

// objectCacheSize bounds the decoded objects kept in memory while pushing
// (go-git default: 96 MiB)
const objectCacheSize = 16 * cache.MiByte

func open(repoUrl string) (*git.Repository, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// EnableDotGitCommonDir: in a linked worktree (.git file), objects and
	// references live in the main repository
	repo, err := git.PlainOpenWithOptions(strings.TrimPrefix(repoUrl, "file://"), &git.PlainOpenOptions{
		EnableDotGitCommonDir: true,
	})
	if err != nil {
		diags.AddError("failed to open repository", fmt.Sprintf("cannot open '%s': %s", repoUrl, err.Error()))
		return nil, diags
	}

	// PlainOpen resolves the .git directory (worktrees, .git files) but loads
	// every object fully in memory: reopen the same directory with bounded
	// memory, a push only needs the object database
	fsStorage, ok := repo.Storer.(*filesystem.Storage)
	if !ok {
		return repo, diags
	}
	st := filesystem.NewStorageWithOptions(
		fsStorage.Filesystem(),
		cache.NewObjectLRU(objectCacheSize),
		filesystem.Options{LargeObjectThreshold: largeObjectThreshold},
	)
	repo, err = git.Open(st, nil)
	if err != nil {
		diags.AddError("failed to open repository", fmt.Sprintf("cannot open '%s': %s", repoUrl, err.Error()))
		return nil, diags
	}

	return repo, diags
}

type CloneOpts func(context.Context, *git.CloneOptions)

func WithBasicAuth(user, password *string) CloneOpts {
	return func(ctx context.Context, co *git.CloneOptions) {
		if user == nil || password == nil {
			tflog.Debug(ctx, "skipping adding auth on this repo")
			return
		}

		tflog.Debug(ctx, "Adding basic auth to clone", map[string]any{
			"user":     user,
			"password": password,
		})
		co.Auth = &http.BasicAuth{Username: *user, Password: *password}
	}
}

func WithCommit(commit *string) CloneOpts {
	return func(ctx context.Context, co *git.CloneOptions) {
		if commit != nil && strings.HasPrefix(*commit, "refs/") {
			co.ReferenceName = plumbing.ReferenceName(*commit)
		}
	}
}

func clone(ctx context.Context, repoUrl string, opts ...CloneOpts) (*git.Repository, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	fs := memory.NewStorage()
	wt := memfs.New()

	cloneOpts := &git.CloneOptions{
		URL:        repoUrl,
		RemoteName: "origin",
		Progress:   os.Stdout,
	}
	for _, opt := range opts {
		opt(ctx, cloneOpts)
	}

	r, err := git.CloneContext(ctx, fs, wt, cloneOpts)
	if err != nil {
		diags.AddError("failed to clone repository", err.Error())
		return nil, diags
	}

	return r, diags
}
