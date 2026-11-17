package tmp

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

// The calls below use routes the public OpenAPI document does not describe yet, so
// the SDK cannot generate them.

// GetAddonsProviders lists the add-on providers of the catalogue and their plans.
// The catalogue is downloaded once per client, see getCatalogue.
func GetAddonsProviders(ctx context.Context, cc *client.Client) client.Response[[]models.AddonProviderInfoFullView] {
	return getCatalogue[[]models.AddonProviderInfoFullView](ctx, cc, "/v2/products/addonproviders")
}

// GetProductInstances lists the application runtimes, those available to ownerID
// when it is set. The catalogue is downloaded once per client and owner, see
// getCatalogue.
func GetProductInstances(ctx context.Context, cc *client.Client, ownerID *string) client.Response[[]AvailableInstanceView] {
	path := "/v2/products/instances"
	if ownerID != nil {
		path += "?for=" + url.QueryEscape(*ownerID)
	}
	return getCatalogue[[]AvailableInstanceView](ctx, cc, path)
}

type catalogueKey struct {
	cc   *client.Client
	path string
}

// catalogueEntry holds the answer of one catalogue route. Its mutex is held
// for the whole download: the callers arriving meanwhile wait for it, then
// read the answer instead of downloading it again.
type catalogueEntry struct {
	mu       sync.Mutex
	response any
}

var catalogues sync.Map // catalogueKey -> *catalogueEntry

// getCatalogue downloads a product catalogue once per client and path, and
// answers from memory afterwards.
//
// Every resource reads a catalogue before it creates anything, so an apply
// creating twenty add-ons downloaded the same 200 KB document twenty times, ten
// at once with the default parallelism. Besides the wasted time, the proxy in
// front of the API ends an HTTP/2 connection carrying several such downloads at
// once (GOAWAY, ENHANCE_YOUR_CALM), which fails every request in flight on it.
//
// A catalogue does not change during a plan or an apply, which is how long a
// provider process lives, so the answer is never refreshed. A failed download
// is not kept: the next caller tries again.
//
// The answer is shared between the callers, they must not modify it.
func getCatalogue[T any](ctx context.Context, cc *client.Client, path string) client.Response[T] {
	value, _ := catalogues.LoadOrStore(catalogueKey{cc: cc, path: path}, &catalogueEntry{})
	entry := value.(*catalogueEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if cached, ok := entry.response.(client.Response[T]); ok {
		return cached
	}

	res := client.Get[T](ctx, cc, path)
	if !res.HasError() {
		entry.response = res
	}
	return res
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
