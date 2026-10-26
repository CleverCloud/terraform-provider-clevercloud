package tmp

import (
	"context"
	"fmt"
	"net/url"

	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

// The calls below use routes the public OpenAPI document does not describe yet, so
// the SDK cannot generate them.

// GetAddonsProviders lists the add-on providers of the catalogue and their plans.
func GetAddonsProviders(ctx context.Context, cc *client.Client) client.Response[[]models.AddonProviderInfoFullView] {
	return client.Get[[]models.AddonProviderInfoFullView](ctx, cc, "/v2/products/addonproviders")
}

// GetProductInstances lists the application runtimes, those available to ownerID
// when it is set.
func GetProductInstances(ctx context.Context, cc *client.Client, ownerID *string) client.Response[[]AvailableInstanceView] {
	path := "/v2/products/instances"
	if ownerID != nil {
		path += "?for=" + url.QueryEscape(*ownerID)
	}
	return client.Get[[]AvailableInstanceView](ctx, cc, path)
}

// ListGithubApplications lists the GitHub repositories the user's linked account can deploy.
func ListGithubApplications(ctx context.Context, cc *client.Client) client.Response[[]OAuthApplicationView] {
	return client.Get[[]OAuthApplicationView](ctx, cc, "/v2/github/applications")
}

// GetPostgreSQL reads a PostgreSQL add-on by its add-on ID.
func GetPostgreSQL(ctx context.Context, cc *client.Client, addonID string) client.Response[PostgresqlAddonView] {
	return client.Get[PostgresqlAddonView](ctx, cc, fmt.Sprintf("/v4/addon-providers/postgresql-addon/addons/%s", url.PathEscape(addonID)))
}

// GetMySQL reads a MySQL add-on by its add-on ID.
func GetMySQL(ctx context.Context, cc *client.Client, addonID string) client.Response[MysqlAddonView] {
	return client.Get[MysqlAddonView](ctx, cc, fmt.Sprintf("/v4/addon-providers/mysql-addon/addons/%s", url.PathEscape(addonID)))
}

// GetMysqlInfos reads the MySQL provider: its versions and shared clusters.
func GetMysqlInfos(ctx context.Context, cc *client.Client) client.Response[ProvidersConsoleView] {
	return client.Get[ProvidersConsoleView](ctx, cc, "/v4/addon-providers/mysql-addon")
}

// GetPulsar reads a Pulsar add-on by its real ID.
func GetPulsar(ctx context.Context, cc *client.Client, pulsarID string) client.Response[Pulsar] {
	return client.Get[Pulsar](ctx, cc, fmt.Sprintf("/v4/addon-providers/addon-pulsar/addons/%s", url.PathEscape(pulsarID)))
}

// GetMateriaKV reads a Materia KV database.
func GetMateriaKV(ctx context.Context, cc *client.Client, ownerID, kvID string) client.Response[MateriaDB] {
	return client.Get[MateriaDB](ctx, cc, fmt.Sprintf("/v4/materia/organisations/%s/materia/databases/%s", url.PathEscape(ownerID), url.PathEscape(kvID)))
}
