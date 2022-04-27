package repository

import (
	"context"
	"testing"
)

func TestRepositoryWithCommit(t *testing.T) {
	ctx := context.Background()
	repository := New()
	expectedSHA := "f4b6aeab4559cc7293249722b956826c3b664076"

	err := repository.Clone(
		ctx,
		"https://github.com/CleverCloud/clever-tools.git",
		"master:"+expectedSHA,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to clone repo: %s", err.Error())
	}

	current, err := repository.Current()
	if err != nil {
		t.Fatalf("failed to clone repo: %s", err.Error())
	}
	if current.Hash().String() != expectedSHA {
		t.Fatalf("current commit does not match, got: %s, expect: %s", current.Hash().String(), expectedSHA)
	}

}

func TestRepositoryWithTag(t *testing.T) {
	ctx := context.Background()
	repository := New()
	expectedSHA := "ce49ad15adfa4121db57ff57efc499a915f3b173"

	err := repository.Clone(
		ctx,
		"https://github.com/CleverCloud/clever-tools.git",
		"2.7.1",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to clone repo: %s", err.Error())
	}

	current, err := repository.Current()
	if err != nil {
		t.Fatalf("failed to clone repo: %s", err.Error())
	}
	if current.Hash().String() != expectedSHA {
		t.Fatalf("current commit does not match, got: %s, expect: %s", current.Hash().String(), expectedSHA)
	}
}
