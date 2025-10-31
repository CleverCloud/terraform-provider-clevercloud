package tmp

import (
	"context"

	"go.clever-cloud.dev/client"
)

// KeycloakInfo represents the complete Keycloak addon information
// as returned by the API endpoint: /v4/addon-providers/addon-keycloak/addons/{keycloak_id}
type KeycloakInfo struct {
	ResourceID         string              `json:"resourceId"`
	AddonID            string              `json:"addonId"`
	Name               string              `json:"name"`
	OwnerID            string              `json:"ownerId"`
	Plan               string              `json:"plan"`
	Version            string              `json:"version"`
	JavaVersion        string              `json:"javaVersion"`
	AccessURL          string              `json:"accessUrl"`
	InitialCredentials KeycloakCredentials `json:"initialCredentials"`
	AvailableVersions  []string            `json:"availableVersions"`
	Resources          KeycloakResources   `json:"resources"`
	Features           KeycloakFeatures    `json:"features"`
	EnvVars            map[string]string   `json:"envVars"`
}

type KeycloakCredentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type KeycloakResources struct {
	Entrypoint string `json:"entrypoint"`
	FsBucketID string `json:"fsbucketId"`
	PgsqlID    string `json:"pgsqlId"`
}

type KeycloakFeatures struct {
	NetworkGroup *string `json:"networkGroup"`
}

// GetKeycloak retrieves Keycloak addon information using the real Keycloak ID
func GetKeycloak(ctx context.Context, cc *client.Client, keycloakID string) client.Response[KeycloakInfo] {
	path := "/v4/addon-providers/addon-keycloak/addons/" + keycloakID
	return client.Get[KeycloakInfo](ctx, cc, path)
}
