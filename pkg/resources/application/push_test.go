package application

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitDeploy_StreamPush(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary required to build the fixture repository")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst.git")
	first := buildLargeRepository(t, src, 1, 4<<10)
	runGit(t, "", "init", "--quiet", "--bare", dst)

	push := func(commit string, wantDeployed bool) {
		t.Helper()

		_, deployed, diags := gitDeploy(context.Background(), Deployment{
			Repository: "file://" + src,
			Commit:     &commit,
		}, "file://"+dst, "")
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
