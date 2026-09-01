package resources

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
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

// NetworkgroupsAttribute is the shared schema for the "networkgroups"
// attribute, used by application runtimes and by add-ons piloting their
// underlying application
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
				Required:            true,
				MarkdownDescription: "domain name which will resolve to application instances inside the networkgroup; must end with a dot followed by one of the platform DNS suffixes served on `GET /networkgroup/configuration` (`cc-ng.cloud` on the public platform)",
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

	ngIDToFQDN := map[string]string{}
	for _, config := range configs {
		ngIDToFQDN[config.NetworkgroupID] = config.FQDN
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

	missingNG := expectedNG.Difference(currentNG)

	// Validate the FQDNs of the members about to be created: the platform
	// rejects a domain name outside its allowed suffixes, and its 400 does not
	// name them. Only members being (re)created are checked, so an existing
	// member with a legacy name never blocks an apply.
	if missingNG.Size() > 0 {
		validateMemberFQDNs(ctx, prov, missingNG.Slice(), ngIDToFQDN, diags)
		if diags.HasError() {
			return
		}
	}

	for ng := range missingNG.Iter() {
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
			detail := addRes.Error().Error()
			var apiErr *client.APIError
			if errors.As(addRes.Error(), &apiErr) && len(apiErr.Context) > 0 {
				// the field-level validation detail only lives in the error context
				detail = fmt.Sprintf("%s %v", apiErr.Message, apiErr.Context)
			}
			diags.AddError("failed to add member to NG", detail)
		}
	}
}

// validateMemberFQDNs checks the FQDN of each member about to be created against
// the deployment's allowed DNS suffixes, fetched from the public
// /networkgroup/configuration route. A deployment predating that route answers
// 404: validation is then skipped and the create call remains the authority.
func validateMemberFQDNs(
	ctx context.Context,
	prov provider.Provider,
	ngIDs []string,
	ngIDToFQDN map[string]string,
	diags *diag.Diagnostics,
) {
	confRes := tmp.GetNetworkgroupConfiguration(ctx, prov.Client())
	if confRes.HasError() {
		tflog.Warn(ctx, "cannot fetch the networkgroup configuration, skipping FQDN validation", map[string]any{"error": confRes.Error().Error()})
		return
	}

	suffixes := confRes.Payload().DNSSuffixes
	if len(suffixes) == 0 {
		return
	}

	for _, ng := range ngIDs {
		fqdn := ngIDToFQDN[ng]
		if !slices.ContainsFunc(suffixes, func(suffix string) bool {
			return strings.HasSuffix(fqdn, "."+suffix)
		}) {
			diags.AddError(
				"invalid networkgroup member fqdn",
				fmt.Sprintf(
					"%q must end with a dot followed by one of the platform DNS suffixes: %s (e.g. %q)",
					fqdn, strings.Join(suffixes, ", "), "myapp."+suffixes[0],
				),
			)
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
				ngObj := types.ObjectValueMust(networkgroupConfigSchema, map[string]attr.Value{
					"networkgroup_id": types.StringValue(ng.ID),
					"fqdn":            types.StringValue(member.DomainName),
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
