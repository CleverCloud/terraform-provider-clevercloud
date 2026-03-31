package oauth_consumer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Available OAuth rights
const (
	RightAccessOrganisations                      = "access_organisations"
	RightAccessOrganisationsBills                 = "access_organisations_bills"
	RightAccessOrganisationsConsumptionStatistics = "access_organisations_consumption_statistics"
	RightAccessOrganisationsCreditCount           = "access_organisations_credit_count"
	RightAccessPersonalInformation                = "access_personal_information"
	RightManageOrganisations                      = "manage_organisations"
	RightManageOrganisationsApplications          = "manage_organisations_applications"
	RightManageOrganisationsMembers               = "manage_organisations_members"
	RightManageOrganisationsServices              = "manage_organisations_services"
	RightManagePersonalInformation                = "manage_personal_information"
	RightManageSSHKeys                            = "manage_ssh_keys"
)

// validRights returns a slice of all valid OAuth rights
func validRights() []string {
	return []string{
		RightAccessOrganisations,
		RightAccessOrganisationsBills,
		RightAccessOrganisationsConsumptionStatistics,
		RightAccessOrganisationsCreditCount,
		RightAccessPersonalInformation,
		RightManageOrganisations,
		RightManageOrganisationsApplications,
		RightManageOrganisationsMembers,
		RightManageOrganisationsServices,
		RightManagePersonalInformation,
		RightManageSSHKeys,
	}
}

// isValidRight checks if a right is in the list of valid rights
func isValidRight(right string) bool {
	for _, validRight := range validRights() {
		if right == validRight {
			return true
		}
	}
	return false
}

// ValidateRights is a validator for the rights Set attribute
func ValidateRights() validator.Set {
	return pkg.NewSetValidator("Validate OAuth rights", func(ctx context.Context, req validator.SetRequest, res *validator.SetResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		items := []types.String{}
		res.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &items, false)...)
		if res.Diagnostics.HasError() {
			return
		}

		for _, item := range items {
			if item.IsUnknown() {
				continue
			}

			right := item.ValueString()
			if !isValidRight(right) {
				res.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid OAuth right",
					fmt.Sprintf("'%s' is not a valid OAuth right. Valid rights are: %v", right, validRights()),
				)
			}
		}
	})
}

// setToRightsRequest converts a Terraform Set to an OAuthConsumerRightsRequest
func (r *ResourceOAuthConsumer) setToRightsRequest(ctx context.Context, rightsSet types.Set, diags *diag.Diagnostics) tmp.OAuthConsumerRightsRequest {
	rights := tmp.OAuthConsumerRightsRequest{}

	if rightsSet.IsNull() || rightsSet.IsUnknown() {
		return rights
	}

	rightsSlice := pkg.SetToStringSlice(ctx, rightsSet, diags)

	// Iterate through the slice and set boolean values based on right name
	for _, right := range rightsSlice {
		switch right {
		case RightAccessOrganisations:
			rights.AccessOrganisations = true
		case RightAccessOrganisationsBills:
			rights.AccessOrganisationsBills = true
		case RightAccessOrganisationsConsumptionStatistics:
			rights.AccessOrganisationsConsumptionStatistics = true
		case RightAccessOrganisationsCreditCount:
			rights.AccessOrganisationsCreditCount = true
		case RightAccessPersonalInformation:
			rights.AccessPersonalInformation = true
		case RightManageOrganisations:
			rights.ManageOrganisations = true
		case RightManageOrganisationsApplications:
			rights.ManageOrganisationsApplications = true
		case RightManageOrganisationsMembers:
			rights.ManageOrganisationsMembers = true
		case RightManageOrganisationsServices:
			rights.ManageOrganisationsServices = true
		case RightManagePersonalInformation:
			rights.ManagePersonalInformation = true
		case RightManageSSHKeys:
			rights.ManageSSHKeys = true
		}
	}

	return rights
}

// rightsResponseToSet converts an OAuthConsumerRightsResponse to a Terraform Set
func (r *ResourceOAuthConsumer) rightsResponseToSet(ctx context.Context, rightsResp tmp.OAuthConsumerRightsResponse, diags *diag.Diagnostics) types.Set {
	var rightsSlice []string

	// Add each enabled right to the slice
	if rightsResp.AccessOrganisations {
		rightsSlice = append(rightsSlice, RightAccessOrganisations)
	}
	if rightsResp.AccessOrganisationsBills {
		rightsSlice = append(rightsSlice, RightAccessOrganisationsBills)
	}
	if rightsResp.AccessOrganisationsConsumptionStatistics {
		rightsSlice = append(rightsSlice, RightAccessOrganisationsConsumptionStatistics)
	}
	if rightsResp.AccessOrganisationsCreditCount {
		rightsSlice = append(rightsSlice, RightAccessOrganisationsCreditCount)
	}
	if rightsResp.AccessPersonalInformation {
		rightsSlice = append(rightsSlice, RightAccessPersonalInformation)
	}
	if rightsResp.ManageOrganisations {
		rightsSlice = append(rightsSlice, RightManageOrganisations)
	}
	if rightsResp.ManageOrganisationsApplications {
		rightsSlice = append(rightsSlice, RightManageOrganisationsApplications)
	}
	if rightsResp.ManageOrganisationsMembers {
		rightsSlice = append(rightsSlice, RightManageOrganisationsMembers)
	}
	if rightsResp.ManageOrganisationsServices {
		rightsSlice = append(rightsSlice, RightManageOrganisationsServices)
	}
	if rightsResp.ManagePersonalInformation {
		rightsSlice = append(rightsSlice, RightManagePersonalInformation)
	}
	if rightsResp.ManageSSHKeys {
		rightsSlice = append(rightsSlice, RightManageSSHKeys)
	}

	return pkg.FromSetString(rightsSlice, diags)
}
