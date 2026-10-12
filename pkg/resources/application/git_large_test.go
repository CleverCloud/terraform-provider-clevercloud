package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const mib = 1 << 20

// TestGitDeploy_LargeRepositoryBoundedMemory reproduces #469: pushing a large
// local repository must not hold the whole repository in memory.
//
// The source repository holds one commit of random (incompressible, so no
// delta can be found) files and is pushed through gitDeploy to a local bare
// repository. Two layouts are pushed: a few large files (streamed from disk
// above the large object threshold) and thousands of small files, like a
// vendored dependency tree. The receiving git-receive-pack runs out of process, so the heap
// measured here is only the provider side of the push.
//
// Heavy test, opt-in:
//
//		LARGE_REPO_MB=128 go test -run TestGitDeploy_LargeRepositoryBoundedMemory ./pkg/resources/application/
//
//	  - LARGE_REPO_MB: repository size (MiB), for each layout
//	  - LARGE_REPO_MAX_HEAP_MB: accepted heap growth during the push (MiB, default 64)
//	  - LARGE_REPO_ABORT_HEAP_MB: hard ceiling over the starting heap, the test process panics above it (MiB, default 4*size)
//	  - LARGE_REPO_DIR: where to build the repositories (default: user cache dir,
//	    t.TempDir() is avoided because /tmp is often a tmpfs, i.e. RAM)
func TestGitDeploy_LargeRepositoryBoundedMemory(t *testing.T) {
	sizeMB := envInt(t, "LARGE_REPO_MB", 0)
	if sizeMB == 0 {
		t.Skip("set LARGE_REPO_MB to run this test")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary required to build the fixture repository")
	}
	maxHeapMB := envInt(t, "LARGE_REPO_MAX_HEAP_MB", 64)
	abortHeapMB := envInt(t, "LARGE_REPO_ABORT_HEAP_MB", 4*sizeMB)

	for _, layout := range []struct {
		name     string
		fileSize int64
	}{
		{"large files", 4 * mib},
		{"thousands of small files", 8 << 10},
	} {
		t.Run(layout.name, func(t *testing.T) {
			testPushBoundedMemory(t, sizeMB, layout.fileSize, maxHeapMB, abortHeapMB)
		})
	}
}

func testPushBoundedMemory(t *testing.T, sizeMB int, fileSize int64, maxHeapMB, abortHeapMB int) {
	base := os.Getenv("LARGE_REPO_DIR")
	if base == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			t.Fatal(err)
		}
		base = cache
	}
	dir, err := os.MkdirTemp(base, "tf-clevercloud-large-repo-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst.git")
	commit := buildLargeRepository(t, src, sizeMB, fileSize)
	runGit(t, "", "init", "--quiet", "--bare", dst)

	// Make the GC follow the live heap closely, so the peak reflects what the
	// push retains and not how lazily garbage is collected
	defer debug.SetGCPercent(debug.SetGCPercent(10))
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var peak atomic.Uint64
	stop := watchHeap(&peak, before.HeapAlloc+uint64(abortHeapMB)*mib)

	_, deployed, diags := gitDeploy(context.Background(), Deployment{
		Repository: "file://" + src,
		Commit:     &commit,
	}, "file://"+dst, "")
	stop()

	if diags.HasError() {
		t.Fatalf("push failed: %v", diags)
	}
	if !deployed {
		t.Fatal("expected the push to trigger a deployment")
	}
	if got := strings.TrimSpace(runGit(t, dst, "rev-parse", "refs/heads/master")); got != commit {
		t.Fatalf("remote master = %s, want %s", got, commit)
	}

	growthMB := (int64(peak.Load()) - int64(before.HeapAlloc)) / mib
	t.Logf("repository: %d MiB in %d files, heap before push: %d MiB, peak heap growth during push: %d MiB",
		sizeMB, int64(sizeMB)*mib/fileSize, before.HeapAlloc/mib, growthMB)
	if growthMB > int64(maxHeapMB) {
		t.Errorf("push retained %d MiB of heap for a %d MiB repository (limit %d MiB): memory scales with the repository size",
			growthMB, sizeMB, maxHeapMB)
	}
}

// buildLargeRepository creates a repository with a single commit made of
// random files of fileSize bytes totalling sizeMB, spread over directories of
// 1000 files, and returns the commit hash.
func buildLargeRepository(t *testing.T, path string, sizeMB int, fileSize int64) string {
	t.Helper()

	runGit(t, "", "init", "--quiet", "--initial-branch=master", path)

	total := int64(sizeMB) * mib
	for i := 0; int64(i)*fileSize < total; i++ {
		sub := filepath.Join(path, fmt.Sprintf("dir-%03d", i/1000))
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(sub, fmt.Sprintf("blob-%04d.bin", i)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.CopyN(f, rand.Reader, min(fileSize, total-int64(i)*fileSize)); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	runGit(t, path, "add", ".")
	runGit(t, path, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "--quiet", "-m", "large")
	// a packed repository, like a real one after `git gc` or a clone (without
	// delta search, so building the fixture stays cheap in memory)
	runGit(t, path, "-c", "pack.threads=1", "repack", "-a", "-d", "-q", "--window=0")

	return strings.TrimSpace(runGit(t, path, "rev-parse", "HEAD"))
}

// watchHeap records the peak heap allocation until the returned function is
// called, and panics above abortAt so a regression cannot take the host down.
func watchHeap(peak *atomic.Uint64, abortAt uint64) func() {
	done := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		var m runtime.MemStats
		for {
			runtime.ReadMemStats(&m)
			if m.HeapAlloc > peak.Load() {
				peak.Store(m.HeapAlloc)
			}
			if m.HeapAlloc > abortAt {
				panic(fmt.Sprintf("heap reached %d MiB (abort ceiling %d MiB), stopping before the host runs out of memory",
					m.HeapAlloc/mib, abortAt/mib))
			}

			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()

	return func() {
		close(done)
		<-stopped
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func envInt(t *testing.T, key string, def int) int {
	t.Helper()

	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		t.Fatalf("%s: %v", key, err)
	}
	return i
}
