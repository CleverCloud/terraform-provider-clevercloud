package application

import (
	"context"
	"testing"

	"go.clever-cloud.dev/client"
)

func Test_gitDeploy(t *testing.T) {

	t.Run("main", func(t *testing.T) {

		ctx := context.Background()
		commit := "refs/heads/toto"
		//commit := "ad3df007f292b301f9b725ebf96e54c7a5dbb4f2"
		cc := client.New(client.WithAutoOauthConfig())
		got := gitDeploy(
			ctx,
			Deployment{
				Repository: "https://github.com/gnoireaux/special__empty",
				Commit:     &commit,
			},
			cc,
			"https://push-n3-par-clevercloud-customers.services.clever-cloud.com/app_9a72c7d9-3efe-4525-a543-c6e676b0124d",
		)

		if got.HasError() {
			t.Errorf("gitDeploy() , diag %v", got)
		}
	})

}
