package networkgroup

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceNG) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceNG.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}
}

// Create a new resource
func (r *ResourceNG) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Networkgroup](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := tmp.GenID()
	fmt.Printf("############## Networkgroup ID: %s\n", id)
	plan.ID = basetypes.NewStringValue(id)
	ngRes := tmp.CreateNetworkgroup(ctx, r.cc, r.org, tmp.NetworkgroupCreation{
		Label:       plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tags:        pkg.SetToStringSlice(ctx, plan.Tags, &resp.Diagnostics),
	})

	fmt.Printf("Networkgroup created: \n%+v\n", ngRes.Payload())

	r.WaitForNG(ctx, r.cc, r.org, id)

	//ng := ngRes.Payload()
	/*plan.ID = basetypes.NewStringValue(ng.ID)
	plan.Name = basetypes.NewStringValue(ng.Label)
	plan.Description = basetypes.NewStringPointerValue(ng.Description)
	plan.Tags = pkg.FromSetString(ng.Tags, &resp.Diagnostics)
	plan.Network = pkg.FromStr(ng.NetworkIP)*/

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceNG) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceNG READ", map[string]any{"request": req})

	state := helper.StateFrom[Networkgroup](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ngRes := tmp.GetNetworkgroup(ctx, r.cc, r.org, state.ID.ValueString())
	if ngRes.HasError() {
		resp.Diagnostics.AddError("failed to get networkgroup", ngRes.Error().Error())
		return
	}
	ng := ngRes.Payload()

	state.Name = basetypes.NewStringValue(ng.Label)
	state.Description = basetypes.NewStringPointerValue(ng.Description)
	state.Tags = pkg.FromSetString(ng.Tags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceNG) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceNG) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Networkgroup DELETE")

	_ = helper.StateFrom[Networkgroup](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

// Import resource
func (r *ResourceNG) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}

func (r *ResourceNG) WaitForNG(ctx context.Context, cc *client.Client, organisationID, ngId string) {
	for range 5 {
		res := tmp.SearchNetworkgroup(ctx, cc, organisationID, ngId)
		if res.HasError() {
			fmt.Printf("ERR: %+v\n", res.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Printf("%+v\n", res.Payload())

	}

}
