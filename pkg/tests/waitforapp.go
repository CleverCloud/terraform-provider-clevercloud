package tests

import (
	"context"
	"fmt"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// WaitForAppInstanceUp polls the application's instances until at least one
// reaches the "UP" state, or the timeout expires.
//
// Use this as a ConfigStateCheck after creating an app whose downstream effects
// (e.g. NetworkGroup peers) only materialize once the runtime is actually running.
func WaitForAppInstanceUp(resourceFullName string, timeout time.Duration) statecheck.StateCheck {
	return waitForAppInstanceUp{
		resourceFullName: resourceFullName,
		timeout:          timeout,
	}
}

type waitForAppInstanceUp struct {
	resourceFullName string
	timeout          time.Duration
}

func (w waitForAppInstanceUp) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	r := pkg.First(req.State.Values.RootModule.Resources, func(sr *tfjson.StateResource) bool {
		return sr.Address == w.resourceFullName
	})
	if r == nil {
		resp.Error = fmt.Errorf("resource %q not found in state", w.resourceFullName)
		return
	}
	idI, ok := (*r).AttributeValues["id"]
	if !ok {
		resp.Error = fmt.Errorf("resource %q has no id attribute", w.resourceFullName)
		return
	}
	appID, ok := idI.(string)
	if !ok {
		resp.Error = fmt.Errorf("resource %q id is not a string", w.resourceFullName)
		return
	}

	cc := client.New(client.WithAutoOauthConfig())
	deadline, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	for {
		select {
		case <-deadline.Done():
			resp.Error = fmt.Errorf("timeout waiting for app %s instance UP: %w", appID, deadline.Err())
			return
		default:
			res := tmp.ListInstances(deadline, cc, ORGANISATION, appID)
			if res.HasError() {
				tflog.Warn(deadline, "ListInstances error", map[string]any{"err": res.Error().Error()})
				time.Sleep(2 * time.Second)
				continue
			}
			for _, inst := range *res.Payload() {
				if inst.State == "UP" {
					tflog.Info(deadline, "app instance is UP", map[string]any{"app": appID, "instance": inst.ID})
					return
				}
			}
			tflog.Info(deadline, "still waiting for app instance UP", map[string]any{"app": appID})
			time.Sleep(2 * time.Second)
		}
	}
}
