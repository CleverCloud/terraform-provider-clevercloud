// This package should be replaced by a well generated one from OpenAPI spec
package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

type AddonRequest struct {
	Name       string            `json:"name"`
	Plan       string            `json:"plan"`
	Options    map[string]string `json:"options"`
	ProviderID string            `json:"providerId"`
	Region     string            `json:"region"`
}

type AddonResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	RealID       string    `json:"realId"`
	Region       string    `json:"region"`
	Plan         AddonPlan `json:"plan"`
	CreationDate int64     `json:"creationDate"`
	ConfigKeys   []string  `json:"configKeys"`
}

type AddonPlan struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AddonProvider struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Plans []AddonPlan `json:"plans"`
}

type PostgreSQL struct {
	// app_id:addon_5abaf3ea-d53f-4021-9711-cd294d50c662
	// creation_date:2022-04-20T08:24:07.28Z[UTC]
	Database string `json:"database" example:"bwf32ifhr5cofspgzrbb"`
	// features:[map[enabled:false name:encryption]]
	Host string `json:"host" example:"bwf32ifhr5cofspgzrbb-postgresql.services.clever-cloud.com"`
	// id:ea97919f-983b-4699-a673-2ed0668bf196
	// owner_id:user_32114ae3-1716-4aa7-8d16-e664ca6ccd1f
	Password string `json:"password" example:"omEbGQw628gIxHK9Bp8d"`
	Plan     string `json:"plan" example:"xs_med"`
	Port     int    `json:"port" example:"6388"`
	// read_only_users:[]
	Status string `json:"status" example:"ACTIVE"`
	User   string `json:"user" example:"uxw1ikwnp6gflbgp5iun"`
	// version:14
	Zone string `json:"zone" example:"par"`
}

func GetAddonsProviders(ctx context.Context, cc *client.Client) client.Response[[]AddonProvider] {
	return client.Get[[]AddonProvider](ctx, cc, "/v2/products/addonproviders")
}

func CreateAddon(ctx context.Context, cc *client.Client, organisation string, addon AddonRequest) client.Response[AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons", organisation)
	return client.Post[AddonResponse](ctx, cc, path, addon)
}

func GetPostgreSQL(ctx context.Context, cc *client.Client, postgresqlID string) client.Response[PostgreSQL] {
	path := fmt.Sprintf("/v4/addon-providers/postgresql-addon/addons/%s", postgresqlID)
	return client.Get[PostgreSQL](ctx, cc, path)
}

func DeletePostgres(ctx context.Context, cc *client.Client, organisationID string, postgresID string) client.Response[interface{}] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s", organisationID, postgresID)
	return client.Delete[interface{}](ctx, cc, path)
}
