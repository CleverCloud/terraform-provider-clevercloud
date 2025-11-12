package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type NetworkgroupConfig struct {
	NetworkgroupID string `tfsdk:"networkgroup_id"`
	FQDN           string `tfsdk:"fqdn"`
}

func SyncNetworkGroups(
	ctx context.Context,
	cc *client.Client,
	kind, orgID, applicationID string, configs []NetworkgroupConfig,
	diags *diag.Diagnostics,
) {
	expectedNG := set.New(pkg.Map(
		configs,
		func(member NetworkgroupConfig) string { return member.NetworkgroupID },
	)...)

	allngRes := tmp.ListNetworkgroups(ctx, cc, orgID)
	if allngRes.HasError() {
		diags.AddError("failed to list Networkgroups", allngRes.Error().Error())
		return
	}
	allNG := *allngRes.Payload()

	_currentNg := pkg.Reduce(allNG, []string{}, func(acc []string, ng tmp.Networkgroup) []string {
		for _, member := range ng.Members {
			if member.ID == applicationID {
				return append(acc, ng.ID)
			}
		}
		return acc
	})
	currentNG := set.New(_currentNg...)

	ngIDToFQDN := map[string]string{}
	for _, config := range configs {
		ngIDToFQDN[config.NetworkgroupID] = config.FQDN
	}

	for inPlaceNG := range expectedNG.Intersection(currentNG).Iter() {
		// a member for this app exists on the expected NG
		memberRes := tmp.GetMember(ctx, cc, orgID, inPlaceNG, applicationID)
		if memberRes.HasError() {
			diags.AddError("failed to get member", memberRes.Error().Error())
			continue
		}

		if memberRes.Payload().DomainName == ngIDToFQDN[inPlaceNG] {
			return
		}

		tflog.Warn(ctx, "a member exists on the expected NG but with an old FQDN, recreate it")
		deleteRes := tmp.DeleteMember(ctx, cc, orgID, inPlaceNG, applicationID)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to remove member from NG", deleteRes.Error().Error())
			continue
		}

		// remove it from set for recreation
		currentNG.Remove(inPlaceNG)
	}

	for ng := range currentNG.Difference(expectedNG).Iter() {
		// app is not in this NG anymore
		deleteRes := tmp.DeleteMember(ctx, cc, orgID, ng, applicationID)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to remove member from ng", deleteRes.Error().Error())
		}
		tflog.Info(ctx, "removed member from NG")
	}

	for ng := range expectedNG.Difference(currentNG).Iter() {
		addRes := tmp.AddMemberToNetworkgroup(ctx, cc, orgID, ng, tmp.Member{
			ID:         applicationID,
			Kind:       kind,
			DomainName: ngIDToFQDN[ng],
		})
		if addRes.HasError() {
			diags.AddError("failed to add member to NG", addRes.Error().Error())
		}
	}
}

// ReadNetworkGroups reads the current networkgroups for a resource and returns them as a types.Set
func ReadNetworkGroups(
	ctx context.Context,
	cc *client.Client,
	orgID, resourceID string,
	diags *diag.Diagnostics,
) types.Set {
	schema := map[string]attr.Type{
		"networkgroup_id": types.StringType,
		"fqdn":            types.StringType,
	}
	null := types.SetNull(types.ObjectType{AttrTypes: schema})

	allngRes := tmp.ListNetworkgroups(ctx, cc, orgID)
	if allngRes.HasError() {
		diags.AddError("failed to list Networkgroups", allngRes.Error().Error())
		return null
	}
	allNG := *allngRes.Payload()

	var networkgroups []attr.Value
	for _, ng := range allNG {
		for _, member := range ng.Members {
			if member.ID == resourceID {
				ngObj := types.ObjectValueMust(schema, map[string]attr.Value{
					"networkgroup_id": types.StringValue(ng.ID),
					"fqdn":            types.StringValue(member.DomainName),
				})
				networkgroups = append(networkgroups, ngObj)
				break
			}
		}
	}
	if len(networkgroups) == 0 {
		return null
	}

	result, d := types.SetValue(types.ObjectType{AttrTypes: schema}, networkgroups)
	diags.Append(d...)
	return result
}
