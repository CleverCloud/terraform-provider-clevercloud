package application

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func commitFile(t *testing.T, repoPath, name, content string) plumbing.Hash {
	t.Helper()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, name), []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := wt.Add(name); err != nil {
		t.Fatalf("add: %v", err)
	}
	hash, err := wt.Commit(content, &git.CommitOptions{
		Author: &object.Signature{Name: "t", Email: "t@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	return hash
}

func TestResolveHeadCommit(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "repo")
	if _, err := git.PlainInit(repoPath, false); err != nil {
		t.Fatalf("init: %v", err)
	}
	c1 := commitFile(t, repoPath, "a", "v1")

	t.Run("local repository HEAD", func(t *testing.T) {
		head, diags := ResolveHeadCommit(t.Context(), "file://"+repoPath, nil, nil)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if head != c1.String() {
			t.Fatalf("expected %s, got %s", c1, head)
		}
	})

	t.Run("local repository HEAD follows new commits", func(t *testing.T) {
		c2 := commitFile(t, repoPath, "a", "v2")
		head, diags := ResolveHeadCommit(t.Context(), "file://"+repoPath, nil, nil)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if head != c2.String() {
			t.Fatalf("expected %s, got %s", c2, head)
		}
	})

	t.Run("remote repository is listed, not cloned", func(t *testing.T) {
		// a plain path goes through the go-git "file" transport, like a remote
		head, diags := ResolveHeadCommit(t.Context(), repoPath, nil, nil)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		repo, _ := git.PlainOpen(repoPath)
		ref, _ := repo.Head()
		if head != ref.Hash().String() {
			t.Fatalf("expected %s, got %s", ref.Hash(), head)
		}
	})

	t.Run("missing repository is an error", func(t *testing.T) {
		_, diags := ResolveHeadCommit(t.Context(), "file://"+filepath.Join(t.TempDir(), "nope"), nil, nil)
		if !diags.HasError() {
			t.Fatal("expected an error")
		}
	})
}
