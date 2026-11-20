package tmp

import (
	"context"
	"fmt"
	"net/url"

	"go.clever-cloud.dev/client"
)

// The calls below use routes the public OpenAPI document either does not
// describe, or describes without the schema of their answer, so the SDK cannot
// generate them with a typed result.

// GetPostgreSQL reads a PostgreSQL add-on by its add-on ID.
func GetPostgreSQL(ctx context.Context, cc *client.Client, addonID string) client.Response[PostgresqlAddonView] {
	return client.Get[PostgresqlAddonView](ctx, cc, fmt.Sprintf("/v4/addon-providers/postgresql-addon/addons/%s", url.PathEscape(addonID)))
}

// GetMySQL reads a MySQL add-on by its add-on ID.
func GetMySQL(ctx context.Context, cc *client.Client, addonID string) client.Response[MysqlAddonView] {
	return client.Get[MysqlAddonView](ctx, cc, fmt.Sprintf("/v4/addon-providers/mysql-addon/addons/%s", url.PathEscape(addonID)))
}
