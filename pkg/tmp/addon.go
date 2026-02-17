// This package should be replaced by a well generated one from OpenAPI spec
package tmp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.dev/client"
)

type AddonRequest struct {
	Name       string            `json:"name"`
	Plan       string            `json:"plan"`
	Options    map[string]string `json:"options"`
	ProviderID string            `json:"providerId"`
	Region     string            `json:"region"`
	Version    string            `json:"version,omitempty"`
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
type AddonPlans []AddonPlan

func (provider *AddonProvider) FirstPlan() *AddonPlan {
	if provider == nil {
		return nil
	}
	if len(provider.Plans) == 0 {
		return nil
	}

	return &provider.Plans[0]
}

type AddonProvider struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Plans AddonPlans `json:"plans"`
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

type PostgreSQLBackup struct {
	BackupID     string    `json:"backup_id"`
	EntityID     string    `json:"entity_id"`
	Status       string    `json:"status"`
	CreationDate time.Time `json:"creation_date"`
	DeleteDate   time.Time `json:"delete_at"`
	Filename     string    `json:"filename"`
	DownloadURL  string    `json:"download_url"`
}

func GetPostgreSQLBackups(ctx context.Context, cc *client.Client, organisationID, postgresqlID string) client.Response[[]PostgreSQLBackup] {
	path := fmt.Sprintf("/v2/backups/%s/%s", organisationID, postgresqlID)
	return client.Get[[]PostgreSQLBackup](ctx, cc, path)
}

type MySQL struct {
	// app_id:addon_5abaf3ea-d53f-4021-9711-cd294d50c662
	CreationDate string `json:"creation_date" example:"2024-03-12T13:38:33.313Z[UTC]"`
	Database     string `json:"database" example:"bwf32ifhr5cofspgzrbb"`
	// features:[map[enabled:false name:encryption]]
	Host string `json:"host" example:"bwf32ifhr5cofspgzrbb-postgresql.services.clever-cloud.com"`
	// id:ea97919f-983b-4699-a673-2ed0668bf196
	// owner_id:user_32114ae3-1716-4aa7-8d16-e664ca6ccd1f
	Password      string              `json:"password" example:"omEbGQw628gIxHK9Bp8d"`
	Plan          string              `json:"plan" example:"xs_med"`
	Port          int                 `json:"port" example:"6388"`
	Status        string              `json:"status" example:"ACTIVE"`
	User          string              `json:"user" example:"uxw1ikwnp6gflbgp5iun"`
	Version       string              `json:"version"` // 14
	Zone          string              `json:"zone" example:"par"`
	Features      []MySQLFeature      `json:"features"`
	ReadOnlyUsers []MySQLReadOnlyUser `json:"read_only_users"`
}

type MySQLReadOnlyUser struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func (p MySQL) Uri() string {
	return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", p.User, p.Password, p.Host, p.Port, p.Database)
}

type MySQLFeature struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type Elasticsearch struct {
	ID             string                 `json:"id"`
	AppID          string                 `json:"app_id"`
	Plan           string                 `json:"plan"`
	Zone           string                 `json:"zone"`
	OwnerID        string                 `json:"owner_id"`
	CreationDate   string                 `json:"creation_date"`
	Status         string                 `json:"status"`
	Host           string                 `json:"host"`
	User           string                 `json:"user"`
	Password       string                 `json:"password"`
	ApmHost        *string                `json:"apm_host"`
	ApmUser        string                 `json:"apm_user"`
	ApmPassword    string                 `json:"apm_password"`
	ApmAuthToken   string                 `json:"apm_auth_token"`
	KibanaHost     *string                `json:"kibana_host"`
	KibanaUser     string                 `json:"kibana_user"`
	KibanaPassword string                 `json:"kibana_password"`
	Version        string                 `json:"version"`
	Plugins        []string               `json:"plugins"`
	Backups        ElasticsearchBackups   `json:"backups"`
	Services       []ElasticsearchService `json:"services"`
	Features       []ElasticsearchFeature `json:"features"`
}

type ElasticsearchBackups struct {
	KibanaSnapshotsURL string `json:"kibana_snapshots_url"`
}

type ElasticsearchService struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (e Elasticsearch) Uri() string {
	return fmt.Sprintf("https://%s:%s@%s:9243", e.User, e.Password, e.Host)
}

type ElasticsearchFeature struct {
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

// Use Addon ID
func GetElasticsearch(ctx context.Context, cc *client.Client, elasticsearchID string) client.Response[Elasticsearch] {
	path := fmt.Sprintf("/v4/addon-providers/es-addon/addons/%s", elasticsearchID)
	return client.Get[Elasticsearch](ctx, cc, path)
}

func GetConfigProvider(ctx context.Context, cc *client.Client, configProviderId string) client.Response[ConfigProvider] {
	path := fmt.Sprintf("/v4/addon-providers/config-provider/addons/%s", configProviderId)
	return client.Get[ConfigProvider](ctx, cc, path)
}

type ConfigProvider struct {
	ID             string            `json:"id"`
	OrganisationID string            `json:"ownerId"`
	Environment    map[string]string `json:"environment"`
	Status         string            `json:"status"`
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
	ResourceID         string               `json:"resourceId"`
	AddonID            string               `json:"addonId"`
	Name               string               `json:"name"`
	OwnerID            string               `json:"ownerId"`
	Plan               string               `json:"plan"`
	Version            string               `json:"version"`
	AccessURL          string               `json:"accessUrl"`
	InitialCredentials *MetabaseCredentials `json:"initialCredentials,omitempty"`
	AvailableVersions  []string             `json:"availableVersions"`
	Resources          MetabaseResources    `json:"resources"`
	Features           MetabaseFeatures     `json:"features"`
	EnvVars            map[string]string    `json:"envVars"`
}

type MetabaseCredentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type MetabaseResources struct {
	PostgresID string `json:"postgresId"`
}

type MetabaseFeatures struct {
	NetworkGroup *MetabaseNetworkGroup `json:"networkGroup,omitempty"`
}

type MetabaseNetworkGroup struct {
	NetworkGroupID string `json:"networkGroupId"`
}

// Use real ID
func GetMetabase(ctx context.Context, cc *client.Client, metabaseID string) client.Response[Metabase] {
	path := fmt.Sprintf("/v4/addon-providers/addon-metabase/addons/%s", metabaseID)
	return client.Get[Metabase](ctx, cc, path)
}

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

type Keycloak struct {
	ResourceID         string                     `json:"resourceId"`
	AddonID            string                     `json:"addonId"`
	Name               string                     `json:"name"`
	OwnerID            string                     `json:"ownerId"`
	Plan               string                     `json:"plan"`
	Version            string                     `json:"version"`
	AccessURL          string                     `json:"accessUrl"`
	InitialCredentials KeycloakInitialCredentials `json:"initialCredentials"`
	AvailableVersions  []string                   `json:"availableVersions"`
	Resources          KeycloakResources          `json:"resources"`
	Features           KeycloakFeatures           `json:"features"`
	EnvVars            map[string]string          `json:"envVars"`
}

type KeycloakResources struct {
	PostgresID string `json:"postgresId"`
	FSBucketID string `json:"fsBucketId"`
}

type KeycloakFeatures struct {
	NetworkGroup *KeycloakNetworkGroup `json:"networkGroup,omitempty"`
}

type KeycloakNetworkGroup struct {
	NetworkGroupID string `json:"networkGroupId"`
}

type KeycloakInitialCredentials struct {
	AdminUsername string `json:"user"`
	AdminPassword string `json:"password"`
}

type Matomo struct {
	ResourceID        string            `json:"resourceId"`
	AddonID           string            `json:"addonId"`
	Name              string            `json:"name"`
	OwnerID           string            `json:"ownerId"`
	Plan              string            `json:"plan"`
	Version           string            `json:"version"`
	PhpVersion        string            `json:"phpVersion"`
	AccessURL         string            `json:"accessUrl"`
	AvailableVersions []string          `json:"availableVersions"`
	Resources         MatomoResources   `json:"resources"`
	EnvVars           map[string]string `json:"envVars"`
}

type MatomoResources struct {
	Entrypoint string `json:"entrypoint"`
	MysqlID    string `json:"mysqlId"`
	RedisID    string `json:"redisId"`
}

// Use real ID
func GetMatomo(ctx context.Context, cc *client.Client, matomoID string) client.Response[Matomo] {
	path := fmt.Sprintf("/v4/addon-providers/addon-matomo/addons/%s", matomoID)
	return client.Get[Matomo](ctx, cc, path)
}

type OtoroshiInfo struct {
	ResourceID        string            `json:"resourceId"`
	AddonID           string            `json:"addonId"`
	Name              string            `json:"name"`
	OwnerID           string            `json:"ownerId"`
	Plan              string            `json:"plan"`
	Version           string            `json:"version"`
	AccessURL         string            `json:"accessUrl"`
	API               *OtoroshiAPI      `json:"api"`
	AvailableVersions []string          `json:"availableVersions"`
	Resources         map[string]string `json:"resources"`
	Features          map[string]any    `json:"features"`
	EnvVars           map[string]string `json:"envVars"`
	Initialredentials struct {
		User      string `json:"user"`
		Passsword string `json:"password"`
	} `json:"initialCredentials"`
}
type OtoroshiAPI struct {
	URL string `json:"url"`
}

// Use real ID
func GetOtoroshi(ctx context.Context, cc *client.Client, otoroshiID string) client.Response[OtoroshiInfo] {
	path := fmt.Sprintf("/v4/addon-providers/addon-otoroshi/addons/%s", otoroshiID)
	return client.Get[OtoroshiInfo](ctx, cc, path)
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

func (evs EnvVars) Map() map[string]string {
	m := map[string]string{}
	for _, item := range evs {
		m[item.Name] = item.Value
	}

	return m
}

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

func UpdateConfigProviderEnv(ctx context.Context, cc *client.Client, organisation string, addon string, envVars EnvVars) client.Response[EnvVars] {
	path := fmt.Sprintf("/v4/addon-providers/config-provider/addons/%s/env", addon)
	return client.Put[EnvVars](ctx, cc, path, envVars)
}

func GetConfigProviderEnv(ctx context.Context, cc *client.Client, organisation string, addon string) client.Response[EnvVars] {
	path := fmt.Sprintf("/v4/addon-providers/config-provider/addons/%s/env", addon)
	return client.Get[EnvVars](ctx, cc, path)
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

type MysqlInfos struct {
	DefaultDedicatedVersion string         `json:"defaultDedicatedVersion"`
	ProviderID              string         `json:"providerId"`
	Clusters                []MysqlCluster `json:"clusters"`
	Dedicated               map[string]any `json:"dedicated"`
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

type ElasticsearchInfos struct {
	DefaultDedicatedVersion string                 `json:"defaultDedicatedVersion"`
	ProviderID              string                 `json:"providerId"`
	Clusters                []ElasticsearchCluster `json:"clusters"`
	Dedicated               map[string]any         `json:"dedicated"`
}

type ElasticsearchCluster struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Region   string `json:"zone"`
	Version  string `json:"version"`
	Features []any  `json:"features"`
}

func GetElasticsearchInfos(ctx context.Context, cc *client.Client) client.Response[ElasticsearchInfos] {
	path := "/v4/addon-providers/es-addon"
	return client.Get[ElasticsearchInfos](ctx, cc, path)
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

// FromMySQLReadOnlyUsers converts a slice of MySQLReadOnlyUser structs to a types.List with nested objects
func FromMySQLReadOnlyUsers(users []MySQLReadOnlyUser) types.List {
	objectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"user":     types.StringType,
			"password": types.StringType,
		},
	}

	if len(users) == 0 {
		return types.ListNull(objectType)
	}

	objects := make([]attr.Value, len(users))
	for i, user := range users {
		objects[i] = types.ObjectValueMust(objectType.AttrTypes, map[string]attr.Value{
			"user":     types.StringValue(user.User),
			"password": types.StringValue(user.Password),
		})
	}

	return types.ListValueMust(objectType, objects)
}

type KubernetesInfo struct {
	CreationDate string `json:"creationDate"`
	Description  string `json:"description"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"` // ACTIVE, DELETED, DELETING, DEPLOYING, FAILED
	Tag          string `json:"tag"`
	TenantID     string `json:"tenantId"`
	Version      string `json:"version"`
}

// GetKubernetes retrieves Kubernetes cluster details using the cluster ID
// This now returns the same structure as GetKubernetesCluster (ClusterView)
func GetKubernetes(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[KubernetesInfo] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s", organisationID, clusterID)
	return client.Get[KubernetesInfo](ctx, cc, path)
}

// GetKubeconfig retrieves the kubeconfig file for a Kubernetes cluster
func GetKubeconfig(ctx context.Context, cc *client.Client, organisationID, addonClusterID string) client.Response[client.PlainTextString] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/kubeconfig.yaml", organisationID, addonClusterID)
	return client.Get[client.PlainTextString](ctx, cc, path)
}

// KubernetesCreateRequest represents the request body for creating a Kubernetes cluster
type KubernetesCreateRequest struct {
	Description      string  `json:"description,omitempty"`
	KubeMajorVersion string  `json:"kubeMajorVersion,omitempty"`
	Name             string  `json:"name"`
	NetworkGroupID   *string `json:"networkGroupId,omitempty"`
	Tag              string  `json:"tag,omitempty"`
}

// ClusterView represents a Kubernetes cluster (same as KubernetesCreateResponse)
type ClusterView struct {
	CreationDate string `json:"creationDate"`
	Description  string `json:"description"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"` // ACTIVE, DELETED, DELETING, DEPLOYING, FAILED
	Tag          string `json:"tag"`
	TenantID     string `json:"tenantId"`
	Version      string `json:"version"`
}

// KubernetesCreateResponse is an alias for ClusterView
type KubernetesCreateResponse = ClusterView

// CreateKubernetes creates a new Kubernetes cluster in the specified organization
func CreateKubernetes(ctx context.Context, cc *client.Client, organisationID string, req KubernetesCreateRequest) client.Response[KubernetesCreateResponse] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters", organisationID)
	return client.Post[KubernetesCreateResponse](ctx, cc, path, req)
}

// ListKubernetesClusters lists all Kubernetes clusters in an organization
func ListKubernetesClusters(ctx context.Context, cc *client.Client, organisationID string) client.Response[[]ClusterView] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters", organisationID)
	return client.Get[[]ClusterView](ctx, cc, path)
}

// GetKubernetesCluster retrieves a specific Kubernetes cluster by ID
func GetKubernetesCluster(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[ClusterView] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s", organisationID, clusterID)
	return client.Get[ClusterView](ctx, cc, path)
}

// DeleteKubernetes deletes a Kubernetes cluster in the specified organization
func DeleteKubernetes(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[ClusterView] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s", organisationID, clusterID)
	return client.Delete[ClusterView](ctx, cc, path)
}

// NodeGroupCreationPayload represents the request body for creating a node group
type NodeGroupCreationPayload struct {
	Description     string            `json:"description,omitempty"`
	Flavor          string            `json:"flavor"`          // xs, s, m, l, xl
	Labels          map[string]string `json:"labels"`          // Required
	MaxNodeCount    int32             `json:"maxNodeCount"`    // Required
	MinNodeCount    int32             `json:"minNodeCount"`    // Required
	Name            string            `json:"name"`            // Required
	Tag             string            `json:"tag,omitempty"`   // Optional zone/region
	Taints          string            `json:"taints"`          // Required
	TargetNodeCount int32             `json:"targetNodeCount"` // Required
}

// NodeGroup represents a Kubernetes node group
type NodeGroup struct {
	ClusterID        string            `json:"clusterId"`
	CreatedAt        string            `json:"createdAt"`
	CurrentNodeCount int32             `json:"currentNodeCount"`
	Description      string            `json:"description"`
	Flavor           string            `json:"flavor"`
	ID               string            `json:"id"`
	Labels           map[string]string `json:"labels"`
	MaxNodeCount     int32             `json:"maxNodeCount"`
	MinNodeCount     int32             `json:"minNodeCount"`
	Name             string            `json:"name"`
	Status           string            `json:"status"` // CREATED, DELETED, DEPLOYED, DEPLOYING, FAILED, READY, RESIZING, TERMINATING
	Tag              string            `json:"tag"`
	Taints           string            `json:"taints"`
	TargetNodeCount  int32             `json:"targetNodeCount"`
	UpdatedAt        string            `json:"updatedAt"`
}

// NodeGroupPatchPayload represents the request body for updating a node group
// Same structure as NodeGroupCreationPayload
type NodeGroupPatchPayload = NodeGroupCreationPayload

// CreateNodeGroup creates a new node group in a Kubernetes cluster
func CreateNodeGroup(ctx context.Context, cc *client.Client, organisationID, clusterID string, req NodeGroupCreationPayload) client.Response[NodeGroup] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/node-groups", organisationID, clusterID)
	return client.Post[NodeGroup](ctx, cc, path, req)
}

// ListNodeGroups lists all node groups in a Kubernetes cluster
func ListNodeGroups(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[[]NodeGroup] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/node-groups", organisationID, clusterID)
	return client.Get[[]NodeGroup](ctx, cc, path)
}

// GetNodeGroup retrieves a specific node group by ID
func GetNodeGroup(ctx context.Context, cc *client.Client, organisationID, clusterID, nodeGroupID string) client.Response[NodeGroup] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/node-groups/%s", organisationID, clusterID, nodeGroupID)
	return client.Get[NodeGroup](ctx, cc, path)
}

// UpdateNodeGroup updates a node group in a Kubernetes cluster
func UpdateNodeGroup(ctx context.Context, cc *client.Client, organisationID, clusterID, nodeGroupID string, req NodeGroupPatchPayload) client.Response[NodeGroup] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/node-groups/%s", organisationID, clusterID, nodeGroupID)
	return client.Patch[NodeGroup](ctx, cc, path, req)
}

// DeleteNodeGroup deletes a node group from a Kubernetes cluster
func DeleteNodeGroup(ctx context.Context, cc *client.Client, organisationID, clusterID, nodeGroupID string) client.Response[NodeGroup] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/node-groups/%s", organisationID, clusterID, nodeGroupID)
	return client.Delete[NodeGroup](ctx, cc, path)
}

// KubeconfigPresignedUrlResponse represents the response from getting a presigned URL for kubeconfig
type KubeconfigPresignedUrlResponse struct {
	URL string `json:"url"`
}

// GetKubeconfigPresignedUrl retrieves a presigned URL for downloading the kubeconfig
func GetKubeconfigPresignedUrl(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[KubeconfigPresignedUrlResponse] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/kubeconfig/presigned-url", organisationID, clusterID)
	return client.Get[KubeconfigPresignedUrlResponse](ctx, cc, path)
}

// CsiCephResponse represents the response from enabling CSI Ceph
type CsiCephResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// EnableCsiCeph enables the CSI Ceph driver for a Kubernetes cluster
func EnableCsiCeph(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[CsiCephResponse] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s/csi/ceph", organisationID, clusterID)
	return client.Post[CsiCephResponse](ctx, cc, path, nil)
}
