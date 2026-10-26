package oauth_consumer

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/sdk/models"
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
	return slices.Contains(validRights(), right)
}

// ValidateRights is a validator for the rights Set attribute
func ValidateRights() validator.Set {
	return pkg.NewSetValidator("Validate OAuth rights", func(ctx context.Context, req validator.SetRequest, res *validator.SetResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		items := pkg.SetToStringSlice(ctx, req.ConfigValue, &res.Diagnostics)
		if res.Diagnostics.HasError() {
			return
		}

		for _, right := range items {
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

// setToRightsRequest converts a Terraform Set to the rights of an OAuth consumer
// request. Every right is sent, granted or not, as the API has always received them.
func (r *ResourceOAuthConsumer) setToRightsRequest(ctx context.Context, rightsSet types.Set, diags *diag.Diagnostics) map[string]bool {
	rights := map[string]bool{}
	for _, right := range []string{
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
	} {
		rights[right] = false
	}

	if rightsSet.IsNull() || rightsSet.IsUnknown() {
		return rights
	}

	for _, right := range pkg.SetToStringSlice(ctx, rightsSet, diags) {
		if _, known := rights[right]; known {
			rights[right] = true
		}
	}

	return rights
}

// rightsResponseToSet converts an OAuthConsumerRightsResponse to a Terraform Set
func (r *ResourceOAuthConsumer) rightsResponseToSet(ctx context.Context, rightsResp models.OAuthRightsView, diags *diag.Diagnostics) types.Set {
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
	if rightsResp.ManageSshKeys {
		rightsSlice = append(rightsSlice, RightManageSSHKeys)
	}

	return pkg.FromSetString(rightsSlice, diags)
}
