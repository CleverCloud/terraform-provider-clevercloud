package repository

import (
	"context"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Repository struct {
	storage storage.Storer
	fs      billy.Filesystem
	repo    *git.Repository
}

func New() *Repository {
	return &Repository{
		storage: memory.NewStorage(),
		fs:      memfs.New(),
	}
}

func (r *Repository) Current() (*plumbing.Reference, error) {
	return r.repo.Head()
}

// ref support either "<tag>" or "<branch>:<commit>"
func (r *Repository) Clone(ctx context.Context, url string, ref string, auth transport.AuthMethod) error {
	var err error
	var commitSha string

	refName := plumbing.NewTagReferenceName(ref)

	if parts := strings.Split(ref, ":"); len(parts) == 2 {
		refName = plumbing.NewBranchReferenceName(parts[0])
		commitSha = parts[1]
	}

	r.repo, err = git.CloneContext(ctx, r.storage, r.fs, &git.CloneOptions{
		URL:             url,
		RemoteName:      "origin",
		InsecureSkipTLS: true,
		ReferenceName:   refName,
		SingleBranch:    true,
		Depth:           1,
		Progress:        os.Stdout,
	})
	if err != nil {
		return err
	}

	if commitSha != "" {
		workTree, err := r.repo.Worktree()
		if err != nil {
			return err
		}

		return workTree.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(commitSha),
			Force: true,
		})
	}

	return nil
}

func (r *Repository) AddCleverRemote(url string) error {
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: "clever",
		URLs: []string{url},
	})
	return err
}

func (r *Repository) Push(ctx context.Context, auth transport.AuthMethod) error {
	return r.repo.PushContext(ctx, &git.PushOptions{
		RemoteName:      "clever",
		Auth:            auth,
		Force:           true,
		InsecureSkipTLS: true,
		RefSpecs:        []config.RefSpec{},
	})
}
