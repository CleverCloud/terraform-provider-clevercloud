package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

type Organisation struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	BillingEmail     string      `json:"billingEmail"`
	Address          string      `json:"address"`
	City             string      `json:"city"`
	Zipcode          string      `json:"zipcode"`
	Country          string      `json:"country"`
	Company          string      `json:"company"`
	Vat              string      `json:"VAT"`
	Avatar           string      `json:"avatar"`
	VatState         string      `json:"vatState"`
	CustomerFullName string      `json:"customerFullName"`
	CanPay           bool        `json:"canPay"`
	CleverEnterprise bool        `json:"cleverEnterprise"`
	EmergencyNumber  interface{} `json:"emergencyNumber"`
	CanSEPA          bool        `json:"canSEPA"`
	IsTrusted        bool        `json:"isTrusted"`
}

func GetOrganisation(ctx context.Context, cc *client.Client, organisationID string) client.Response[Organisation] {
	path := fmt.Sprintf("/v2/organisations/%s", organisationID)
	return client.Get[Organisation](ctx, cc, path)
}
