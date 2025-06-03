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
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	RealID       string                `json:"realId"`
	Region       string                `json:"region"`
	Plan         AddonPlan             `json:"plan"`
	Provider     AddonResponseProvider `json:"provider"`
	CreationDate int64                 `json:"creationDate"`
	ConfigKeys   []string              `json:"configKeys"`
}

type AddonResponseProvider struct {
	ID string `json:"id"`
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

type MateriaKV struct {
	ID             string `json:"id"`
	ClusterID      string `json:"clusterId"`
	OrganisationID string `json:"ownerId"`
	Kind           string `json:"kind"`
	Plan           string `json:"plan"`
	Host           string `json:"host"`
	Port           int64  `json:"port"`
	Token          string `json:"token"`
	TokenID        string `json:"tokenId"`
	Status         string `json:"status"`
	// ccapiUrl	"https://api.clever-cloud.com/v2/vendor/apps/addon_dbf12716-9353-41ef-aabf-e4b7fce1ba5e"
}

func GetMateriaKV(ctx context.Context, cc *client.Client, organisationID, postgresqlID string) client.Response[MateriaKV] {
	path := fmt.Sprintf("/v4/materia/organisations/%s/materia/databases/%s", organisationID, postgresqlID)
	return client.Get[MateriaKV](ctx, cc, path)
}

type Metabase struct {
	OwnerID        string                `json:"ownerId"`
	ID             string                `json:"addonId"`
	NetworkgroupID *string               `json:"networkgroupId"`
	PostgresID     string                `json:"postgresId"`
	Status         string                `json:"status" example:"ACTIVE"`
	Applications   []MetabaseApplication `json:"applications"`
}
type MetabaseApplication struct {
	MetabaseID        string `json:"appId"`
	MetabasePlan      string `json:"planIdentifier"`
	Host              string `json:"host"`
	JavaApplicationID string `json:"javaId"`
}

func CreateMetabase(ctx context.Context, cc *client.Client, organisation string, addon AddonRequest) client.Response[AddonResponse] {
	path := "/v2/providers/addon-metabase/resources"
	return client.Post[AddonResponse](ctx, cc, path, addon)
}

func GetMetabase(ctx context.Context, cc *client.Client, metabaseID string) client.Response[Metabase] {
	path := fmt.Sprintf("/v4/addon-providers/addon-metabase/addons/%s", metabaseID)
	return client.Get[Metabase](ctx, cc, path)
}

type MongoDB struct {
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	Status   string `json:"status" example:"ACTIVE"`
	User     string `tfsdk:"user"`
	Password string `tfsdk:"password"`
}

func GetMongoDB(ctx context.Context, cc *client.Client, mongodbID string) client.Response[MongoDB] {
	path := fmt.Sprintf("/v4/addon-providers/mongodb-addon/addons/%s", mongodbID)
	return client.Get[MongoDB](ctx, cc, path)
}

type Keycloak struct {
	OwnerID        string                `json:"ownerId"`
	ID             string                `json:"addonId"`
	NetworkgroupID *string               `json:"networkgroupId"`
	PostgresID     string                `json:"postgresId"`
	FSBucketID     string                `json:"fsBucketId"`
	Applications   []KeycloakApplication `json:"applications"`
}
type KeycloakApplication struct {
	KeycloakID        string `json:"appId"`
	KeycloakPlan      string `json:"planIdentifier"`
	Host              string `json:"host"`
	JavaApplicationID string `json:"javaId"`
}

// Not working ?
func GetKeycloak(ctx context.Context, cc *client.Client, organisationID, keycloakID string) client.Response[Keycloak] {
	path := fmt.Sprintf("v4/keycloaks/organisations/%s/keycloaks/%s", organisationID, keycloakID)
	return client.Get[Keycloak](ctx, cc, path)
}

type DeleteAddonResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func DeleteAddon(ctx context.Context, cc *client.Client, organisationID string, addonID string) client.Response[DeleteAddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s", organisationID, addonID)
	return client.Delete[DeleteAddonResponse](ctx, cc, path)
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type EnvVars []EnvVar

func GetAddon(ctx context.Context, cc *client.Client, organisation string, addon string) client.Response[AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s", organisation, addon)
	return client.Get[AddonResponse](ctx, cc, path)
}

func GetAddonEnv(ctx context.Context, cc *client.Client, organisation string, addon string) client.Response[EnvVars] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s/env", organisation, addon)
	return client.Get[EnvVars](ctx, cc, path)
}

func UpdateAddon(ctx context.Context, cc *client.Client, organisation string, addon string, env map[string]string) client.Response[AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s", organisation, addon)
	return client.Put[AddonResponse](ctx, cc, path, env)
}
