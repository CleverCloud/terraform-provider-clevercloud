// Code generated from go.clever-cloud.dev/sdk v0.3.1. DO NOT EDIT.

package tmp

import (
	"go.clever-cloud.dev/sdk/models"
)

// The models of the routes the public OpenAPI document does not describe yet, as
// the SDK generated them before it moved to that document. They leave with the calls
// of this package once the document publishes those routes.

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

// PostgresqlReadOnlyUserView One read-only user, console-facing spelling (`{user, password}`) — the shape [`PostgresqlAddonVie...
type PostgresqlReadOnlyUserView struct {
	Password string `json:"password"`
	User     string `json:"user"`
}
