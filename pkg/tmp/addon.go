// Package tmp holds the Clever Cloud API calls the provider makes by hand,
// because the OpenAPI document the SDK is generated from does not describe them
// well enough to generate them: the route is missing, its response has no schema,
// or its body is not JSON. Each one should move to go.clever-cloud.dev/sdk once
// the document is fixed.
package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

type MongoDB struct {
	Host     string           `json:"host"`
	Port     int64            `json:"port"`
	Status   string           `json:"status" example:"ACTIVE"`
	User     string           `json:"user"`
	Password string           `json:"password"`
	Database string           `json:"database"`
	Features []MongoDBFeature `json:"features"`
}

type MongoDBFeature struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (mg MongoDB) Uri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", mg.User, mg.Password, mg.Host, mg.Port, mg.Database)
}

// Use Addon ID
func GetMongoDB(ctx context.Context, cc *client.Client, mongodbID string) client.Response[MongoDB] {
	path := fmt.Sprintf("/v4/addon-providers/mongodb-addon/addons/%s", mongodbID)
	return client.Get[MongoDB](ctx, cc, path)
}

type AddonMigrationRequest struct {
	PlanID  string  `json:"planId"`
	Region  string  `json:"region"`
	Version *string `json:"version"`
}

type AddonMigrationResponse struct {
	MigrationID string               `json:"migrationId"`
	RequestDate string               `json:"requestDate"`
	Steps       []AddonMigrationStep `json:"steps"`
	Status      string               `json:"status"`
}

type AddonMigrationStep struct {
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
	Status    string `json:"status,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Message   string `json:"message,omitempty"`
}

func MigrateAddon(ctx context.Context, cc *client.Client, organisationID string, addonID string, req AddonMigrationRequest) client.Response[AddonMigrationResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s/migrations", organisationID, addonID)
	return client.Post[AddonMigrationResponse](ctx, cc, path, req)
}

func ListAddonMigrations(ctx context.Context, cc *client.Client, organisationID string, addonID string) client.Response[[]AddonMigrationResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s/migrations", organisationID, addonID)
	return client.Get[[]AddonMigrationResponse](ctx, cc, path)
}

func GetAddonMigrations(ctx context.Context, cc *client.Client, organisationID, addonID, migrationID string) client.Response[AddonMigrationResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s/migrations/%s", organisationID, addonID, migrationID)
	return client.Get[AddonMigrationResponse](ctx, cc, path)
}
