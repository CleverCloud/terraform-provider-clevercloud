package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk"
	"go.clever-cloud.dev/sdk/models"
)

type NetworkgroupConfig struct {
	NetworkgroupID string `tfsdk:"networkgroup_id"`
	FQDN           string `tfsdk:"fqdn"`
}

var networkgroupConfigSchema = map[string]attr.Type{
	"networkgroup_id": types.StringType,
	"fqdn":            types.StringType,
}
var NullNetworkgroupConfig = types.SetNull(types.ObjectType{AttrTypes: networkgroupConfigSchema})

// networkgroupZone is the DNS zone networkgroup members live in. An older zone,
// ng.clever-cloud.com, is still served for historical members
const networkgroupZone = "cc-ng.cloud"

// CanonicalFQDN builds the domain name the platform assigns to a networkgroup member.
// It is used whenever the configuration leaves fqdn empty, because that value cannot be
// written by hand: it embeds the member ID, which only exists once the resource has been
// created, and Terraform forbids a resource from referring to itself
func CanonicalFQDN(memberID, networkgroupID string) string {
	return fmt.Sprintf("%s.m.%s.%s", memberID, networkgroupID, networkgroupZone)
}

// NetworkgroupsAttribute is the shared schema for the "networkgroups"
// attribute, used by application runtimes and by add-ons
var NetworkgroupsAttribute = schema.SetNestedAttribute{
	Optional:            true,
	MarkdownDescription: "List of networkgroups the application must be part of",
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"networkgroup_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the networkgroup",
			},
			"fqdn": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "domain name which will resolve to application instances inside the networkgroup. When omitted, the platform naming scheme `<memberId>.m.<networkgroupId>.cc-ng.cloud` is used",
			},
		},
	},
}

func SyncNetworkGroups(
	ctx context.Context,
	prov provider.Provider,
	kind, applicationID string, configs []NetworkgroupConfig,
	diags *diag.Diagnostics,
) {
	expectedNG := set.New(pkg.Map(
		configs,
		func(member NetworkgroupConfig) string { return member.NetworkgroupID },
	)...)

	if prov.IsNetwrkgroupsDisabled() {
		tflog.Warn(ctx, "skipping Networkgroups synchronisation because feature is disabled at provider level")
		return
	}

	// Create SDK instance from provider client
	sdkClient := sdk.NewSDK(sdk.WithClient(prov.Client()))

	allngRes := tmp.ListNetworkgroups(ctx, prov.Client(), prov.Organization())
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

	// Resolve the target FQDN once, so that both the "is it up to date" comparison
	// below and the member creation agree on the value an empty fqdn stands for
	ngIDToFQDN := map[string]string{}
	for _, config := range configs {
		fqdn := config.FQDN
		if fqdn == "" {
			fqdn = CanonicalFQDN(applicationID, config.NetworkgroupID)
		}
		ngIDToFQDN[config.NetworkgroupID] = fqdn
	}

	for inPlaceNG := range expectedNG.Intersection(currentNG).Iter() {
		// a member for this app exists on the expected NG
		memberRes := sdkClient.
			V4().
			Networkgroups().
			Organisations().
			Ownerid(prov.Organization()).
			Networkgroups().
			Networkgroupid(inPlaceNG).
			Members().
			Memberid(applicationID).
			Getnetworkgroupmember(ctx)
		if memberRes.HasError() {
			diags.AddError("failed to get member", memberRes.Error().Error())
			continue
		}

		if memberRes.Payload().DomainName == ngIDToFQDN[inPlaceNG] {
			return
		}

		tflog.Warn(ctx, "a member exists on the expected NG but with an old FQDN, recreate it")
		deleteRes := sdkClient.
			V4().
			Networkgroups().
			Organisations().
			Ownerid(prov.Organization()).
			Networkgroups().
			Networkgroupid(inPlaceNG).
			Members().
			Memberid(applicationID).
			Deletenetworkgroupmember(ctx)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to remove member from NG", deleteRes.Error().Error())
			continue
		}

		// remove it from set for recreation
		currentNG.Remove(inPlaceNG)
	}

	for ng := range currentNG.Difference(expectedNG).Iter() {
		// app is not in this NG anymore
		deleteRes := sdkClient.
			V4().
			Networkgroups().
			Organisations().
			Ownerid(prov.Organization()).
			Networkgroups().
			Networkgroupid(ng).
			Members().
			Memberid(applicationID).
			Deletenetworkgroupmember(ctx)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to remove member from ng", deleteRes.Error().Error())
		}
		tflog.Info(ctx, "removed member from NG")
	}

	for ng := range expectedNG.Difference(currentNG).Iter() {
		addRes := sdkClient.
			V4().
			Networkgroups().
			Organisations().
			Ownerid(prov.Organization()).
			Networkgroups().
			Networkgroupid(ng).
			Members().
			Createnetworkgroupmember(ctx, &models.WannabeNetworkgroupMember{
				ID:         applicationID,
				Kind:       models.MemberKind(kind),
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
	prov provider.Provider,
	resourceID string,
	diags *diag.Diagnostics,
) types.Set {

	if prov.IsNetwrkgroupsDisabled() {
		tflog.Warn(ctx, "skipping Networkgroups synchronisation because feature is disabled at provider level")
		return NullNetworkgroupConfig
	}

	allngRes := tmp.ListNetworkgroups(ctx, prov.Client(), prov.Organization())
	if allngRes.HasError() {
		diags.AddError("failed to list Networkgroups", allngRes.Error().Error())
		return NullNetworkgroupConfig
	}
	allNG := *allngRes.Payload()

	var networkgroups []attr.Value
	for _, ng := range allNG {
		for _, member := range ng.Members {
			if member.ID == resourceID {
				// A member carrying the canonical name is one we named ourselves from an
				// empty fqdn, report it back as null so the configuration keeps matching
				fqdn := types.StringValue(member.DomainName)
				if member.DomainName == CanonicalFQDN(resourceID, ng.ID) {
					fqdn = types.StringNull()
				}

				ngObj := types.ObjectValueMust(networkgroupConfigSchema, map[string]attr.Value{
					"networkgroup_id": types.StringValue(ng.ID),
					"fqdn":            fqdn,
				})
				networkgroups = append(networkgroups, ngObj)
				break
			}
		}
	}
	if len(networkgroups) == 0 {
		return NullNetworkgroupConfig
	}

	result, d := types.SetValue(types.ObjectType{AttrTypes: networkgroupConfigSchema}, networkgroups)
	diags.Append(d...)
	return result
}
