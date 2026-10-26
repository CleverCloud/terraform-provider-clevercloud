// Code generated from go.clever-cloud.dev/sdk v0.3.1. DO NOT EDIT.

package tmp

import (
	"time"

	"go.clever-cloud.dev/sdk/models"
)

// The models of the routes the public OpenAPI document does not describe yet, as
// the SDK generated them before it moved to that document. They leave with the calls
// of this package once the document publishes those routes.

// AvailableInstanceView AvailableInstanceView — matches cc-api openapi/ccapiv3.json AvailableInstanceView schema.  Thirte...
type AvailableInstanceView struct {
	BuildFlavor   *models.Flavor             `json:"buildFlavor,omitempty"`
	ComingSoon    bool                       `json:"comingSoon"`
	DefaultFlavor *models.Flavor             `json:"defaultFlavor,omitempty"`
	Deployments   []string                   `json:"deployments"`
	Description   string                     `json:"description"` // Nullable, as `ccapidto AvailableInstanceView.java:17` declares it: a NULL `instance_types.descrip...
	Enabled       bool                       `json:"enabled"`
	Flavors       []models.Flavor            `json:"flavors"`
	MaxInstances  models.InstanceCount       `json:"maxInstances"`
	Name          string                     `json:"name"`
	Tags          []models.SemanticTag       `json:"tags"`
	Type          string                     `json:"type"`
	Variant       models.InstanceVariantView `json:"variant"`
	Version       string                     `json:"version"`
}

// MysqlAddonView `GET /v4/.../addons/{appId}` and internal getAddon body. Field names are the legacy snake_case (N...
type MysqlAddonView struct {
	AppID         string             `json:"app_id"`
	CreationDate  *string            `json:"creation_date,omitempty"`
	Database      string             `json:"database"`
	DeletionDate  *string            `json:"deletion_date,omitempty"`
	Features      []models.Feature   `json:"features"`
	Host          string             `json:"host"`
	ID            string             `json:"id"`
	OwnerID       *string            `json:"owner_id,omitempty"`
	Password      string             `json:"password"`
	Plan          *MysqlPlanType     `json:"plan,omitempty"`
	Port          int                `json:"port"`
	ReadOnlyUsers []ReadOnlyUserView `json:"read_only_users"`
	Status        *MysqlDbStatus     `json:"status,omitempty"`
	User          string             `json:"user"`
	Version       string             `json:"version"`
	Zone          string             `json:"zone"`
}

// ProvidersConsoleView `GET /v4/addon-providers/mysql-addon` body.  Source: AddonProvider.scala — ProvidersConsoleView
type ProvidersConsoleView struct {
	Clusters                []MysqlClusterConsole           `json:"clusters"`
	Dedicated               map[string]models.DedicatedView `json:"dedicated"`
	DefaultDedicatedVersion string                          `json:"defaultDedicatedVersion"`
	ProviderID              string                          `json:"providerId"`
}

// OAuthApplicationView A third-party repository an application can be created from — one entry of `GET /v2/github/applic...
type OAuthApplicationView struct {
	DefaultBranch *string `json:"defaultBranch,omitempty"`
	Description   *string `json:"description,omitempty"`
	GitURL        *string `json:"gitUrl,omitempty"`
	ID            string  `json:"id"` // GitHub's numeric repository id, as a string.
	Name          string  `json:"name"`
	Owner         string  `json:"owner"`
	Priv          bool    `json:"priv"`
}

// PostgresqlAddonView The addon as `GET .../addons/{appId}` (internal AND `/v4`) render it — reference `postgresql_addo...
type PostgresqlAddonView struct {
	AppID         string                       `json:"app_id"`
	CreationDate  string                       `json:"creation_date"`
	Database      string                       `json:"database"`
	DeletionDate  *string                      `json:"deletion_date,omitempty"`
	Features      []PostgresqlFeatureFlag      `json:"features"`
	Host          string                       `json:"host"`
	ID            string                       `json:"id"`
	Locale        *string                      `json:"locale,omitempty"` // The locale recorded at provisioning (`provision_database.locale`, `en_GB` when none was), as `ll_...
	OwnerID       *string                      `json:"owner_id,omitempty"`
	Password      string                       `json:"password"`
	Plan          string                       `json:"plan"`
	Port          int                          `json:"port"`
	ReadOnlyUsers []PostgresqlReadOnlyUserView `json:"read_only_users"`
	Status        string                       `json:"status"`
	User          string                       `json:"user"`
	Version       string                       `json:"version"`
	Zone          string                       `json:"zone"`
}

// Pulsar A Pulsar add-on: its tenant and namespace, cluster, access token, lifecycle dates and cold-storag...
type Pulsar struct {
	AskForDeletionDate        *time.Time   `json:"ask_for_deletion_date,omitempty"` // When deletion was requested
	ClusterID                 string       `json:"cluster_id"`                      // Cluster ID the addon is deployed on
	ColdStorageID             *string      `json:"cold_storage_id,omitempty"`       // Cellar cold storage addon ID (if configured)
	ColdStorageLinked         bool         `json:"cold_storage_linked"`             // Whether cold storage offload is linked
	ColdStorageMustBeProvided bool         `json:"cold_storage_must_be_provided"`   // Whether cold storage must be provisioned for this plan
	CreationDate              time.Time    `json:"creation_date"`                   // Creation timestamp
	DeletionDate              *time.Time   `json:"deletion_date,omitempty"`         // When deletion was completed
	ID                        string       `json:"id"`                              // Provider-internal resource ID (pulsar_*)
	Namespace                 string       `json:"namespace"`                       // Pulsar namespace name
	OwnerID                   string       `json:"owner_id"`                        // Owner (user or org) ID
	Plan                      string       `json:"plan"`                            // Plan identifier
	Status                    PulsarStatus `json:"status"`                          // Addon lifecycle status
	Tenant                    string       `json:"tenant"`                          // Pulsar tenant name
	Token                     string       `json:"token"`                           // Biscuit authentication token (base64url)
}

// MateriaDB A Materia KV database: its cluster, owner, plan, connection host and port, access token and status.
type MateriaDB struct {
	ClusterID string `json:"clusterId"`
	Host      string `json:"host"`
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	OwnerID   string `json:"ownerId"`
	Plan      string `json:"plan"`
	Port      int    `json:"port"`
	Status    string `json:"status"`
	Token     string `json:"token"` // Access token of the database; it is also the password of its Redis-compatible endpoint.
	TokenID   string `json:"tokenId"`
}

// MysqlPlanType represents The `plan_type` DB enum — 34 values, exactly the production enum (which has NO `free` value even ...
type MysqlPlanType string

const (
	MysqlPlanTypeDev        MysqlPlanType = "dev"
	MysqlPlanTypeS          MysqlPlanType = "s"
	MysqlPlanTypeM          MysqlPlanType = "m"
	MysqlPlanTypeMl         MysqlPlanType = "ml"
	MysqlPlanTypeL          MysqlPlanType = "l"
	MysqlPlanTypeXl         MysqlPlanType = "xl"
	MysqlPlanTypeXxl        MysqlPlanType = "xxl"
	MysqlPlanTypeLm         MysqlPlanType = "lm"
	MysqlPlanTypeSimbiose   MysqlPlanType = "simbiose"
	MysqlPlanTypeXsSml      MysqlPlanType = "xs_sml"
	MysqlPlanTypeXsMed      MysqlPlanType = "xs_med"
	MysqlPlanTypeXsBig      MysqlPlanType = "xs_big"
	MysqlPlanTypeSSml       MysqlPlanType = "s_sml"
	MysqlPlanTypeSMed       MysqlPlanType = "s_med"
	MysqlPlanTypeSBig       MysqlPlanType = "s_big"
	MysqlPlanTypeMSml       MysqlPlanType = "m_sml"
	MysqlPlanTypeMMed       MysqlPlanType = "m_med"
	MysqlPlanTypeMBig       MysqlPlanType = "m_big"
	MysqlPlanTypeLSml       MysqlPlanType = "l_sml"
	MysqlPlanTypeLMed       MysqlPlanType = "l_med"
	MysqlPlanTypeLBig       MysqlPlanType = "l_big"
	MysqlPlanTypeXlSml      MysqlPlanType = "xl_sml"
	MysqlPlanTypeXlMed      MysqlPlanType = "xl_med"
	MysqlPlanTypeXlBig      MysqlPlanType = "xl_big"
	MysqlPlanTypeXxlSml     MysqlPlanType = "xxl_sml"
	MysqlPlanTypeXxlMed     MysqlPlanType = "xxl_med"
	MysqlPlanTypeXxlBig     MysqlPlanType = "xxl_big"
	MysqlPlanTypeXxlHug     MysqlPlanType = "xxl_hug"
	MysqlPlanTypeMajDigital MysqlPlanType = "maj_digital"
	MysqlPlanTypeXxsSml     MysqlPlanType = "xxs_sml"
	MysqlPlanTypeXxsMed     MysqlPlanType = "xxs_med"
	MysqlPlanTypeXxsBig     MysqlPlanType = "xxs_big"
	MysqlPlanTypeXsTny      MysqlPlanType = "xs_tny"
	MysqlPlanType_3xlCpuTny MysqlPlanType = "3xl_cpu_tny"
)

// String returns the underlying string value
func (e MysqlPlanType) String() string {
	return string(e)
}

// MysqlDbStatus represents The `addon_status` DB enum.  Schema: addon-api-mysql/conf/evolutions/default/1.sql — CREATE TYPE ...
type MysqlDbStatus string

const (
	MysqlDbStatusACTIVE        MysqlDbStatus = "ACTIVE"
	MysqlDbStatusToDelete      MysqlDbStatus = "TO_DELETE"
	MysqlDbStatusDELETED       MysqlDbStatus = "DELETED"
	MysqlDbStatusQuotaExceeded MysqlDbStatus = "QUOTA_EXCEEDED"
)

// String returns the underlying string value
func (e MysqlDbStatus) String() string {
	return string(e)
}

// ReadOnlyUserView A read-only user view `{user, password}`.  Source: addonprovider-lib Migration.scala — ReadOnlyUs...
type ReadOnlyUserView struct {
	Password string `json:"password"`
	User     string `json:"user"`
}

// PostgresqlFeatureFlag A named option and whether it is on, as V1 renders it.  Source: rewrite/rust/src/provider/views.r...
type PostgresqlFeatureFlag struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name"`
}

// MysqlClusterConsole `ClusterConsoleView { id, label, zone, version, features: [Feature] }` (v4 provider view; `featur...
type MysqlClusterConsole struct {
	Features []models.Feature `json:"features"`
	ID       string           `json:"id"`
	Label    string           `json:"label"`
	Version  string           `json:"version"`
	Zone     string           `json:"zone"`
}

// PostgresqlReadOnlyUserView One read-only user, console-facing spelling (`{user, password}`) — the shape [`PostgresqlAddonVie...
type PostgresqlReadOnlyUserView struct {
	Password string `json:"password"`
	User     string `json:"user"`
}

// PulsarStatus represents Lifecycle status of a Pulsar add-on: `ACTIVE`, `TO_DELETE` (deletion requested), `DELETED`, `NAME...
type PulsarStatus string

const (
	PulsarStatusACTIVE             PulsarStatus = "ACTIVE"
	PulsarStatusToDelete           PulsarStatus = "TO_DELETE"
	PulsarStatusDELETED            PulsarStatus = "DELETED"
	PulsarStatusNamespaceDeleted   PulsarStatus = "NAMESPACE_DELETED"
	PulsarStatusColdStorageDeleted PulsarStatus = "COLD_STORAGE_DELETED"
)

// String returns the underlying string value
func (e PulsarStatus) String() string {
	return string(e)
}
