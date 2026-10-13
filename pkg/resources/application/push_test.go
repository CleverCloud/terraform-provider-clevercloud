package application

import (
	"context"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestGitDeploy_StreamPush(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary required to build the fixture repository")
	}

	// file:// pipes the pack to git-receive-pack, http:// goes through the
	// streamed HTTP request, like the Clever Cloud remote
	for _, transport := range []string{"file", "http"} {
		t.Run(transport, func(t *testing.T) {
			testStreamPush(t, transport)
		})
	}
}

func testStreamPush(t *testing.T, transport string) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst.git")
	first := buildLargeRepository(t, src, 1, 4<<10)
	runGit(t, "", "init", "--quiet", "--bare", dst)
	remote := "file://" + dst
	if transport == "http" {
		remote = serveGitHTTP(t, dir) + "/dst.git"
	}

	push := func(commit string, wantDeployed bool) {
		t.Helper()

		_, deployed, diags := gitDeploy(context.Background(), Deployment{
			Repository: "file://" + src,
			Commit:     &commit,
		}, remote, "")
		if diags.HasError() {
			t.Fatalf("push failed: %v", diags)
		}
		if deployed != wantDeployed {
			t.Fatalf("deployed = %t, want %t", deployed, wantDeployed)
		}
		if got := strings.TrimSpace(runGit(t, dst, "rev-parse", "refs/heads/master")); got != commit {
			t.Fatalf("remote master = %s, want %s", got, commit)
		}
		runGit(t, dst, "fsck", "--strict", "--no-dangling")
	}

	// first push: the whole history
	push(first, true)

	// second push: only the objects the remote does not have yet
	if err := os.WriteFile(filepath.Join(src, "dir-000", "blob-0000.bin"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, src, "add", ".")
	runGit(t, src, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "--quiet", "-m", "second")
	second := strings.TrimSpace(runGit(t, src, "rev-parse", "HEAD"))
	push(second, true)

	// force push back to the first commit, then nothing left to push
	push(first, true)
	push(first, false)
}

// a linked worktree (.git file) keeps its objects in the main repository
func TestGitDeploy_StreamPushFromWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary required to build the fixture repository")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	wt := filepath.Join(dir, "wt")
	dst := filepath.Join(dir, "dst.git")
	commit := buildLargeRepository(t, src, 1, 4<<10)
	runGit(t, src, "worktree", "add", "--quiet", "--detach", wt, commit)
	runGit(t, "", "init", "--quiet", "--bare", dst)

	_, deployed, diags := gitDeploy(context.Background(), Deployment{
		Repository: "file://" + wt,
		Commit:     &commit,
	}, "file://"+dst, "")
	if diags.HasError() {
		t.Fatalf("push failed: %v", diags)
	}
	if !deployed {
		t.Fatal("expected the push to trigger a deployment")
	}
	if got := strings.TrimSpace(runGit(t, dst, "rev-parse", "refs/heads/master")); got != commit {
		t.Fatalf("remote master = %s, want %s", got, commit)
	}
}

// serveGitHTTP serves the bare repositories of root over git's smart HTTP
// protocol (git http-backend, pushes allowed) and returns the server URL.
func serveGitHTTP(t *testing.T, root string) string {
	t.Helper()

	srv := httptest.NewServer(gitHTTPBackend(t, root))
	t.Cleanup(srv.Close)

	return srv.URL
}

func gitHTTPBackend(t *testing.T, root string) http.Handler {
	t.Helper()

	execPath := strings.TrimSpace(runGit(t, "", "--exec-path"))
	backend := &cgi.Handler{
		Path: filepath.Join(execPath, "git-http-backend"),
		Env: []string{
			"GIT_PROJECT_ROOT=" + root,
			"GIT_HTTP_EXPORT_ALL=1",
			"REMOTE_USER=test", // receive-pack is only enabled for authenticated users
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// net/http/cgi rejects chunked bodies, which the streamed push (and
		// git itself above http.postBuffer) sends: like Apache or nginx, pass
		// the body without CONTENT_LENGTH, git http-backend reads it to EOF
		r.TransferEncoding = nil
		backend.ServeHTTP(w, r)
	})
}

// a push rejected because of its size gets the same answer when retried
func TestGitDeploy_TooLargeNotRetried(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary required to build the fixture repository")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	commit := buildLargeRepository(t, src, 1, 4<<10)
	runGit(t, "", "init", "--quiet", "--bare", filepath.Join(dir, "dst.git"))

	backend := gitHTTPBackend(t, dir)
	var pushes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/git-receive-pack") {
			pushes.Add(1)
			http.Error(w, "too large", http.StatusRequestEntityTooLarge)
			return
		}
		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	diags := diag.Diagnostics{}
	GitDeploy(context.Background(), &Deployment{
		Repository: "file://" + src,
		Commit:     &commit,
	}, srv.URL+"/dst.git", "", &diags)

	if !diags.HasError() {
		t.Fatal("expected the push to fail")
	}
	if got := diags.Errors()[0].Summary(); got != "repository too large for the clever remote" {
		t.Errorf("error summary = %q", got)
	}
	if got := pushes.Load(); got != 1 {
		t.Errorf("push attempts = %d, want 1", got)
	}
}
