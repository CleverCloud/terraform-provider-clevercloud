package tests

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type CheckRemoteResource[T any] struct {
	resourceFullName string
	fetch            func(ctx context.Context, id string) (*T, error)
	check            func(ctx context.Context, id string, state *tfjson.State, resource *T) error
}

func (c CheckRemoteResource[T]) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	r := pkg.First(req.State.Values.RootModule.Resources, func(stateResource *tfjson.StateResource) bool {
		return stateResource.Address == c.resourceFullName
	})
	if r == nil {
		resp.Error = fmt.Errorf("resource '%s' not found in state", c.resourceFullName)
		return
	}
	resource := *r

	idI, ok := resource.AttributeValues["id"]
	if !ok {
		resp.Error = fmt.Errorf("resource '%s' does not have 'id' attribute", c.resourceFullName)
		return
	}

	id, ok := idI.(string)
	if !ok {
		resp.Error = fmt.Errorf("resource '%s' 'id' attribute is not a string", c.resourceFullName)
		return
	}

	t, err := c.fetch(ctx, id)
	if err != nil {
		resp.Error = err
		return
	}
	resp.Error = c.check(ctx, id, req.State, t)
}

func NewCheckRemoteResource[T any](
	resourceFullName string,
	fetch func(ctx context.Context, id string) (*T, error),
	check func(ctx context.Context, id string, state *tfjson.State, resource *T) error) *CheckRemoteResource[T] {
	return &CheckRemoteResource[T]{
		resourceFullName: resourceFullName,
		fetch:            fetch,
		check:            check,
	}
}
