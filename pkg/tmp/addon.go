// Package tmp holds the Clever Cloud API calls the provider makes by hand,
// because the OpenAPI document the SDK is generated from does not describe them
// well enough to generate them: the route is missing, its response has no schema,
// or its body is not JSON. Each one should move to go.clever-cloud.dev/sdk once
// the document is fixed.
package tmp

import (
	"context"
	"fmt"
	"time"

	"go.clever-cloud.dev/client"
)

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

// Use Addon ID
func GetElasticsearch(ctx context.Context, cc *client.Client, elasticsearchID string) client.Response[Elasticsearch] {
	path := fmt.Sprintf("/v4/addon-providers/es-addon/addons/%s", elasticsearchID)
	return client.Get[Elasticsearch](ctx, cc, path)
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

type ElasticsearchInfos struct {
	DefaultDedicatedVersion string                         `json:"defaultDedicatedVersion"`
	ProviderID              string                         `json:"providerId"`
	Clusters                []ElasticsearchProviderCluster `json:"clusters"`
	Dedicated               map[string]any                 `json:"dedicated"`
}

// ElasticsearchProviderCluster is a shared cluster of the es-addon provider,
// unrelated to the dedicated ElasticsearchCluster product.
type ElasticsearchProviderCluster struct {
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

type KubernetesInfo struct {
	CreationDate string              `json:"creationDate"`
	Description  string              `json:"description"`
	Features     *KubernetesFeatures `json:"features,omitempty"`
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Status       string              `json:"status"` // ACTIVE, DELETED, DELETING, DEPLOYING, FAILED, RECONCILING
	Tags         []string            `json:"tags"`
	TenantID     string              `json:"tenantId"`
	Topology     *KubernetesTopology `json:"topologyConfig,omitempty"`
	Version      string              `json:"version"`
	ZoneID       string              `json:"zoneId"`
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

// KubernetesFeatures represents the optional features of a Kubernetes cluster.
// Fields are pointers so "absent" and "false" stay distinguishable: the features
// patch is a merge patch, an omitted field is left untouched by the API while an
// explicit false disables the feature
type KubernetesFeatures struct {
	Csi                  *bool `json:"csi,omitempty"`
	NodeAutoprovisioning *bool `json:"nodeAutoprovisioning,omitempty"`
}

// KubernetesTopology is the control plane topology of a cluster. ALL_IN_ONE and
// DEDICATED_COMPUTE take a flavor and a replication factor, DISTRIBUTED takes
// both for each component instead
type KubernetesTopology struct {
	Topology          string                        `json:"topology"`
	Flavor            string                        `json:"flavor,omitempty"`
	ReplicationFactor int64                         `json:"replicationFactor,omitempty"`
	Components        *KubernetesTopologyComponents `json:"components,omitempty"`
}

// KubernetesTopologyComponents holds the settings of each control plane
// component of a DISTRIBUTED topology
type KubernetesTopologyComponents struct {
	Apiserver              KubernetesTopologyComponent `json:"apiserver"`
	CloudControllerManager KubernetesTopologyComponent `json:"cloudControllerManager"`
	ControllerManager      KubernetesTopologyComponent `json:"controllerManager"`
	NodeGroupOperator      KubernetesTopologyComponent `json:"nodeGroupOperator"`
	Scheduler              KubernetesTopologyComponent `json:"scheduler"`
}

// KubernetesTopologyComponent is the size of one control plane component
type KubernetesTopologyComponent struct {
	Flavor            string `json:"flavor"`
	ReplicationFactor int64  `json:"replicationFactor"`
}

// KubernetesCreateRequest represents the request body for creating a Kubernetes cluster
type KubernetesCreateRequest struct {
	Description    string              `json:"description,omitempty"`
	Features       *KubernetesFeatures `json:"features,omitempty"`
	Name           string              `json:"name"`
	NetworkGroupID *string             `json:"networkGroupId,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
	Topology       *KubernetesTopology `json:"topologyConfig,omitempty"`
	Version        string              `json:"version,omitempty"`
	Zone           string              `json:"zone,omitempty"`
}

// KubernetesPatchRequest represents the request body for updating a Kubernetes
// cluster, only the fields set here are modified
type KubernetesPatchRequest struct {
	Features *KubernetesFeatures `json:"features,omitempty"`
	Name     *string             `json:"name,omitempty"`
	// a pointer so an empty list, which removes every tag, is still sent
	Tags *[]string `json:"tags,omitempty"`
}

// KubernetesWaiting tells what a cluster creation is waiting for, the API only
// reports it during that wait
type KubernetesWaiting struct {
	Attempts   int    `json:"attempts"`
	Dependency string `json:"dependency"`
	Since      string `json:"since"`
	Step       string `json:"step,omitempty"`
}

// ClusterView represents a Kubernetes cluster (same as KubernetesCreateResponse)
type ClusterView struct {
	CreationDate string              `json:"creationDate"`
	Description  string              `json:"description"`
	Features     *KubernetesFeatures `json:"features,omitempty"`
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Status       string              `json:"status"` // ACTIVE, DELETED, DELETING, DEPLOYING, FAILED, RECONCILING
	Tags         []string            `json:"tags"`
	TenantID     string              `json:"tenantId"`
	Topology     *KubernetesTopology `json:"topologyConfig,omitempty"`
	Version      string              `json:"version"`
	ZoneID       string              `json:"zoneId"`
	Waiting      *KubernetesWaiting  `json:"waiting,omitempty"`
}

// KubernetesCreateResponse is an alias for ClusterView
type KubernetesCreateResponse = ClusterView

// CreateKubernetes creates a new Kubernetes cluster in the specified organization
func CreateKubernetes(ctx context.Context, cc *client.Client, organisationID string, req KubernetesCreateRequest) client.Response[KubernetesCreateResponse] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters", organisationID)
	return client.Post[KubernetesCreateResponse](ctx, cc, path, req)
}

// KubernetesZone is a zone an organisation may create a cluster in
type KubernetesZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListKubernetesZones lists the zones the organisation may create a cluster in
func ListKubernetesZones(ctx context.Context, cc *client.Client, organisationID string) client.Response[[]KubernetesZone] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/zones", organisationID)
	return client.Get[[]KubernetesZone](ctx, cc, path)
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

// UpdateKubernetes patches a Kubernetes cluster, only the fields set in the
// payload are modified, the others are left untouched by the API
func UpdateKubernetes(ctx context.Context, cc *client.Client, organisationID, clusterID string, req KubernetesPatchRequest) client.Response[ClusterView] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s", organisationID, clusterID)
	return client.Patch[ClusterView](ctx, cc, path, req)
}

// DeleteKubernetes deletes a Kubernetes cluster in the specified organization
func DeleteKubernetes(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[ClusterView] {
	path := fmt.Sprintf("/v4/kubernetes/organisations/%s/clusters/%s", organisationID, clusterID)
	return client.Delete[ClusterView](ctx, cc, path)
}
