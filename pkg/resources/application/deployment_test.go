package application

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

const (
	testSHA        = "a397296e135b24e682a011e31f8e15f2fa8a5a0e"
	testAnotherSHA = "b5f4de2a1c1b24e682a011e31f8e15f2fa8a5a0e"
)

func TestResolveUnknownCommit(t *testing.T) {
	t.Run("nil deployment is a no-op", func(t *testing.T) {
		resolveUnknownCommit(nil, testSHA)
	})

	t.Run("unknown commit resolved to the deployed one", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringUnknown()}
		resolveUnknownCommit(d, testSHA)
		if d.Commit.ValueString() != testSHA {
			t.Fatalf("expected %s, got %s", testSHA, d.Commit)
		}
	})

	t.Run("unknown commit resolved to null when nothing was deployed", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringUnknown()}
		resolveUnknownCommit(d, "")
		if !d.Commit.IsNull() {
			t.Fatalf("expected null, got %s", d.Commit)
		}
	})

	t.Run("known commit is kept", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringValue("refs/heads/main")}
		resolveUnknownCommit(d, testSHA)
		if d.Commit.ValueString() != "refs/heads/main" {
			t.Fatalf("expected refs/heads/main, got %s", d.Commit)
		}
	})
}

func TestSyncDeploymentCommit(t *testing.T) {
	t.Run("nil deployment is a no-op", func(t *testing.T) {
		syncDeploymentCommit(nil, testSHA)
	})

	t.Run("null commit takes the running one", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringNull()}
		syncDeploymentCommit(d, testSHA)
		if d.Commit.ValueString() != testSHA {
			t.Fatalf("expected %s, got %s", testSHA, d.Commit)
		}
	})

	t.Run("hash commit takes the running one (drift)", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringValue(testSHA)}
		syncDeploymentCommit(d, testAnotherSHA)
		if d.Commit.ValueString() != testAnotherSHA {
			t.Fatalf("expected %s, got %s", testAnotherSHA, d.Commit)
		}
	})

	t.Run("git reference is kept", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringValue("refs/heads/main")}
		syncDeploymentCommit(d, testSHA)
		if d.Commit.ValueString() != "refs/heads/main" {
			t.Fatalf("expected refs/heads/main, got %s", d.Commit)
		}
	})

	t.Run("github_hook is kept", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringValue(attributes.GITHUB_COMMIT_PREFIX)}
		syncDeploymentCommit(d, testSHA)
		if d.Commit.ValueString() != attributes.GITHUB_COMMIT_PREFIX {
			t.Fatalf("expected %s, got %s", attributes.GITHUB_COMMIT_PREFIX, d.Commit)
		}
	})

	t.Run("empty running commit keeps the state", func(t *testing.T) {
		d := &attributes.Deployment{Commit: types.StringNull()}
		syncDeploymentCommit(d, "")
		if !d.Commit.IsNull() {
			t.Fatalf("expected null, got %s", d.Commit)
		}
	})
}
