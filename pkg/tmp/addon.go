// This package should be replaced by a well generated one from OpenAPI spec
package tmp

import (
	"context"
	"fmt"
	"strings"

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
	CreationDate string `json:"creation_date" example:"2024-03-12T13:38:33.313Z[UTC]"`
	Database     string `json:"database" example:"bwf32ifhr5cofspgzrbb"`
	// features:[map[enabled:false name:encryption]]
	Host string `json:"host" example:"bwf32ifhr5cofspgzrbb-postgresql.services.clever-cloud.com"`
	// id:ea97919f-983b-4699-a673-2ed0668bf196
	// owner_id:user_32114ae3-1716-4aa7-8d16-e664ca6ccd1f
	Password string `json:"password" example:"omEbGQw628gIxHK9Bp8d"`
	Plan     string `json:"plan" example:"xs_med"`
	Port     int    `json:"port" example:"6388"`
	// read_only_users:[]
	Status   string              `json:"status" example:"ACTIVE"`
	User     string              `json:"user" example:"uxw1ikwnp6gflbgp5iun"`
	Version  string              `json:"version"` // 14
	Zone     string              `json:"zone" example:"par"`
	Features []PostgreSQLFeature `json:"features"`
}

func (p PostgreSQL) Uri() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", p.User, p.Password, p.Host, p.Port, p.Database)
}

type PostgreSQLFeature struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}


type MySQL struct {
	// app_id:addon_5abaf3ea-d53f-4021-9711-cd294d50c662
	CreationDate string `json:"creation_date" example:"2024-03-12T13:38:33.313Z[UTC]"`
	Database     string `json:"database" example:"bwf32ifhr5cofspgzrbb"`
	// features:[map[enabled:false name:encryption]]
	Host string `json:"host" example:"bwf32ifhr5cofspgzrbb-postgresql.services.clever-cloud.com"`
	// id:ea97919f-983b-4699-a673-2ed0668bf196
	// owner_id:user_32114ae3-1716-4aa7-8d16-e664ca6ccd1f
	Password string `json:"password" example:"omEbGQw628gIxHK9Bp8d"`
	Plan     string `json:"plan" example:"xs_med"`
	Port     int    `json:"port" example:"6388"`
	// read_only_users:[]
	Status   string              `json:"status" example:"ACTIVE"`
	User     string              `json:"user" example:"uxw1ikwnp6gflbgp5iun"`
	Version  string              `json:"version"` // 14
	Zone     string              `json:"zone" example:"par"`
	Features []MySQLFeature `json:"features"`
}

func (p MySQL) Uri() string {
	return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", p.User, p.Password, p.Host, p.Port, p.Database)
}

type MySQLFeature struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func GetAddonsProviders(ctx context.Context, cc *client.Client) client.Response[[]AddonProvider] {
	return client.Get[[]AddonProvider](ctx, cc, "/v2/products/addonproviders")
}

func CreateAddon(ctx context.Context, cc *client.Client, organisation string, addon AddonRequest) client.Response[AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons", organisation)
	return client.Post[AddonResponse](ctx, cc, path, addon)
}

// Use Addon ID
func GetPostgreSQL(ctx context.Context, cc *client.Client, postgresqlID string) client.Response[PostgreSQL] {
	path := fmt.Sprintf("/v4/addon-providers/postgresql-addon/addons/%s", postgresqlID)
	return client.Get[PostgreSQL](ctx, cc, path)
}

// Use Addon ID
func GetMySQL(ctx context.Context, cc *client.Client, mysqlID string) client.Response[MySQL] {
	path := fmt.Sprintf("/v4/addon-providers/mysql-addon/addons/%s", mysqlID)
	return client.Get[MySQL](ctx, cc, path)
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

// Use real ID
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
	Database string `tfsdk:"database"`
}


func (mg MongoDB) Uri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", mg.User, mg.Password, mg.Host, mg.Port, mg.Database)
}

// Use Addon ID
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
	path := fmt.Sprintf("/v4/keycloaks/organisations/%s/keycloaks/%s", organisationID, keycloakID)
	return client.Get[Keycloak](ctx, cc, path)
}

type OtoroshiInfo struct {
	ResourceID            string            `json:"resourceId"`
	AddonID               string            `json:"addonId"`
	Name                  string            `json:"name"`
	OwnerID               string            `json:"ownerId"`
	Plan                  string            `json:"plan"`
	Version               string            `json:"version"`
	AccessURL             string            `json:"accessUrl"`
	API                   *OtoroshiAPI      `json:"api"`
	AvailableVersions     []string          `json:"availableVersions"`
	Resources             map[string]string `json:"resources"`
	Features              map[string]any    `json:"features"`
	EnvVars               map[string]string `json:"envVars"`
	APIClientID           string            // Extracted from envVars
	APIClientSecret       string            // Extracted from envVars
	APIURL                string            // Extracted from envVars
	InitialAdminLogin     string            // Extracted from envVars
	InitialAdminPassword  string            // Extracted from envVars
	URL                   string            // Extracted from envVars
}

type OtoroshiAPI struct {
	URL string `json:"url"`
}

// Use real ID
func GetOtoroshi(ctx context.Context, cc *client.Client, otoroshiID string) client.Response[OtoroshiInfo] {
	path := fmt.Sprintf("/v4/addon-providers/addon-otoroshi/addons/%s", otoroshiID)
	resp := client.Get[OtoroshiInfo](ctx, cc, path)

	// Extract specific env vars to dedicated fields if response is successful
	if !resp.HasError() && resp.Payload() != nil {
		info := resp.Payload()
		if info.EnvVars != nil {
			info.APIClientID = info.EnvVars["CC_OTOROSHI_API_CLIENT_ID"]
			info.APIClientSecret = info.EnvVars["CC_OTOROSHI_API_CLIENT_SECRET"]
			info.APIURL = info.EnvVars["CC_OTOROSHI_API_URL"]
			info.InitialAdminLogin = info.EnvVars["CC_OTOROSHI_INITIAL_ADMIN_LOGIN"]
			info.InitialAdminPassword = info.EnvVars["CC_OTOROSHI_INITIAL_ADMIN_PASSWORD"]
			info.URL = info.EnvVars["CC_OTOROSHI_URL"]
		}
	}

	return resp
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

func ListAddons(ctx context.Context, cc *client.Client, organisation string) client.Response[[]AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/addons", organisation)
	return client.Get[[]AddonResponse](ctx, cc, path)
}

type PostgresInfos struct {
	DefaultDedicatedVersion string            `json:"defaultDedicatedVersion"`
	ProviderID              string            `json:"providerId"`
	Clusters                []PostgresCluster `json:"clusters"`
	Dedicated               map[string]any    `json:"dedicated"`
}

type PostgresCluster struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Region   string `json:"zone"`
	Version  string `json:"version"`
	Features []any  `json:"features"`
}

func GetPostgresInfos(ctx context.Context, cc *client.Client) client.Response[PostgresInfos] {
	path := "/v4/addon-providers/postgresql-addon"
	return client.Get[PostgresInfos](ctx, cc, path)
}


type MysqlInfos struct {
	DefaultDedicatedVersion string            `json:"defaultDedicatedVersion"`
	ProviderID              string            `json:"providerId"`
	Clusters                []MysqlCluster    `json:"clusters"`
	Dedicated               map[string]any    `json:"dedicated"`
}

type MysqlCluster struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Region   string `json:"zone"`
	Version  string `json:"version"`
	Features []any  `json:"features"`
}

func GetMysqlInfos(ctx context.Context, cc *client.Client) client.Response[MysqlInfos] {
	path := "/v4/addon-providers/mysql-addon"
	return client.Get[MysqlInfos](ctx, cc, path)
}

func RealIDsToAddonIDs(ctx context.Context, client *client.Client, organisation string, realID ...string) ([]string, error) {
	addonsRes := ListAddons(ctx, client, organisation)
	if addonsRes.HasError() {
		return nil, addonsRes.Error()
	}
	addons := *addonsRes.Payload()

	addonIDs := make([]string, len(realID))
	for i, id := range realID {
		if strings.HasPrefix(id, "addon_") {
			addonIDs[i] = id
			continue
		}

		notFound := true

		for _, addon := range addons {
			if addon.RealID == id {
				addonIDs[i] = addon.ID
				notFound = false
			}
		}

		if notFound {
			return nil, fmt.Errorf("addon %s not found", id)
		}
	}

	return addonIDs, nil
}

func RealIDToAddonID(ctx context.Context, client *client.Client, organisation string, realID string) (string, error) {
	if strings.HasPrefix(realID, "addon_") {
		return realID, nil
	}

	addonsRes := ListAddons(ctx, client, organisation)
	if addonsRes.HasError() {
		return "", addonsRes.Error()
	}
	addons := *addonsRes.Payload()

	for _, addon := range addons {
		if addon.RealID == realID {
			return addon.ID, nil
		}
	}

	return "", fmt.Errorf("addon %s not found", realID)
}

func AddonIDToRealID(ctx context.Context, client *client.Client, organisation string, addonID string) (string, error) {
	if !strings.HasPrefix(addonID, "addon_") {
		return addonID, nil
	}

	addonsRes := ListAddons(ctx, client, organisation)
	if addonsRes.HasError() {
		return "", addonsRes.Error()
	}
	addons := *addonsRes.Payload()

	for _, addon := range addons {
		if addon.ID == addonID {
			return addon.RealID, nil
		}
	}

	return "", fmt.Errorf("addon %s not found", addonID)
}
