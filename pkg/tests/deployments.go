package tests

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

// WaitForDeploymentsSettled polls the application deployments until every
// one of them reached a final state (OK or FAIL) and returns them, most
// recent first.
func WaitForDeploymentsSettled(ctx context.Context, cc *client.Client, appID string, timeout time.Duration) ([]models.DeploymentView, error) {
	deadline, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.Done():
			return nil, fmt.Errorf("timeout waiting for deployments of %s to settle: %w", appID, deadline.Err())
		case <-ticker.C:
			res := pkg.SDK(cc).V2().Organisations().ID(ORGANISATION).Applications().Appid(appID).Deployments().Getapplicationdeploymentsfororga(deadline)
			if res.IsNotFoundError() {
				continue
			}
			if res.HasError() {
				return nil, fmt.Errorf("failed to list deployments: %w", res.Error())
			}

			deployments := *res.Payload()
			if len(deployments) == 0 {
				continue
			}

			settled := true
			for _, d := range deployments {
				if d.State != "OK" && d.State != "FAIL" {
					settled = false
				}
			}
			if !settled {
				continue
			}

			sort.Slice(deployments, func(i, j int) bool { return pkg.Deref(deployments[i].Date) > pkg.Deref(deployments[j].Date) })
			return deployments, nil
		}
	}
}
