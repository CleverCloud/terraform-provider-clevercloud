package registry_test

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/registry"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addonprovider"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/resources/configprovider"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/cellar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/elasticsearch"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/fsbucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/materiakv"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mongodb"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mysql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/postgresql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/pulsar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/redis"
	"go.clever-cloud.com/terraform-provider/pkg/resources/drain"
	"go.clever-cloud.com/terraform-provider/pkg/resources/elasticsearch_cluster"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes/nodegroup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/networkgroup"
	oauthconsumer "go.clever-cloud.com/terraform-provider/pkg/resources/oauth_consumer"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/keycloak"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/matomo"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/metabase"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/otoroshi"
)

// mappedState names the state struct every registered resource fills from the
// API, so that the contract below can be checked by reflection.
//
// A resource missing from this table fails the test. That is the point: it is
// how a new resource, or one whose mappers were renamed away, cannot ship
// without someone looking at its API-to-state mapping.
//
// The 17 application runtimes all share application.Runtime's FromApp and their
// own FromEnv; one entry per runtime would assert the same two methods 17 times,
// so they are covered by the shared entry and listed in runtimeResources.
var mappedState = map[string]any{
	"clevercloud_addon":                 &addon.Addon{},
	"clevercloud_addon_provider":        &addonprovider.AddonProvider{},
	"clevercloud_cellar":                &cellar.Cellar{},
	"clevercloud_configprovider":        &configprovider.ConfigProvider{},
	"clevercloud_elasticsearch":         &elasticsearch.Elasticsearch{},
	"clevercloud_elasticsearch_cluster": &elasticsearch_cluster.ElasticsearchCluster{},
	"clevercloud_fsbucket":              &fsbucket.FSBucket{},
	"clevercloud_keycloak":              &keycloak.Keycloak{},
	"clevercloud_kubernetes":            &kubernetes.Kubernetes{},
	"clevercloud_kubernetes_nodegroup":  &nodegroup.KubernetesNodegroup{},
	"clevercloud_materia_kv":            &materiakv.MateriaKV{},
	"clevercloud_matomo":                &matomo.Matomo{},
	"clevercloud_metabase":              &metabase.Metabase{},
	"clevercloud_mongodb":               &mongodb.MongoDB{},
	"clevercloud_mysql":                 &mysql.MySQL{},
	"clevercloud_networkgroup":          &networkgroup.Networkgroup{},
	"clevercloud_oauth_consumer":        &oauthconsumer.OAuthConsumer{},
	"clevercloud_otoroshi":              &otoroshi.Otoroshi{},
	"clevercloud_postgresql":            &postgresql.PostgreSQL{},
	"clevercloud_pulsar":                &pulsar.Pulsar{},
	"clevercloud_redis":                 &redis.Redis{},

	// The seven drains share one generic resource and one mapper shape.
	"clevercloud_drain_datadog":       &drain.DatadogDrain{},
	"clevercloud_drain_newrelic":      &drain.NewRelicDrain{},
	"clevercloud_drain_elasticsearch": &drain.ElasticsearchDrain{},
	"clevercloud_drain_syslog_udp":    &drain.SyslogUDPDrain{},
	"clevercloud_drain_syslog_tcp":    &drain.SyslogTCPDrain{},
	"clevercloud_drain_http":          &drain.HTTPDrain{},
	"clevercloud_drain_ovh":           &drain.OVHDrain{},
}

// unmapped records the resources that read nothing back from the API, with the
// reason. An entry here is a claim someone has to defend at review, not a way
// to silence the test.
var unmapped = map[string]string{
	// Read performs no API call at all: the bucket is created through the S3
	// endpoint and the platform exposes nothing to read back about it. Giving it
	// a mapper means first giving its Read something to call.
	"clevercloud_cellar_bucket": "no readable API surface",
}

// runtimeResources are the application runtimes, covered by application.Runtime.
var runtimeResources = map[string]bool{
	"clevercloud_docker": true, "clevercloud_dotnet": true, "clevercloud_frankenphp": true,
	"clevercloud_go": true, "clevercloud_haskell": true, "clevercloud_java_war": true,
	"clevercloud_java_jar": true, "clevercloud_java_maven": true, "clevercloud_linux": true,
	"clevercloud_nodejs": true, "clevercloud_php": true, "clevercloud_play2": true,
	"clevercloud_python": true, "clevercloud_ruby": true, "clevercloud_rust": true,
	"clevercloud_scala": true, "clevercloud_static": true, "clevercloud_static_apache": true,
	"clevercloud_v": true,
}

// Every registered resource reads its state back from the API, and must say so
// in one place: a From* method on the state struct it fills.
//
// WHAT THIS PROVES: no resource escapes the convention, and no mapper can be
// renamed or deleted without a reviewer noticing. A new resource has to be added
// to the table above, which is where someone asks whether its Computed
// attributes are actually mapped.
//
// WHAT IT DOES NOT PROVE: that the mappers are called, that they map the right
// field to the right attribute, or that a Computed attribute is never left null.
// Those live in each resource's own mapping_test.go — see CONTRIBUTING.md for
// the three cases every one of them carries. Do not let this test's green
// substitute for them.
func TestResources_everyStateStructHasAMapper(t *testing.T) {
	ctx := t.Context()

	for _, newResource := range registry.Resources {
		r := newResource()

		meta := resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "clevercloud"}, &meta)

		if runtimeResources[meta.TypeName] {
			continue
		}

		if reason, known := unmapped[meta.TypeName]; known {
			t.Logf("%s reads nothing back from the API: %s", meta.TypeName, reason)
			continue
		}

		state, declared := mappedState[meta.TypeName]
		if !declared {
			t.Errorf("%s has no entry in mappedState: add the state struct it fills from the API, "+
				"and check while you are there that its Computed attributes are mapped (see CONTRIBUTING.md)",
				meta.TypeName)
			continue
		}

		if mappers := mappersOf(state); len(mappers) == 0 {
			t.Errorf("%s: %T has no From* mapper", meta.TypeName, state)
		}
	}
}

// The shared runtime mapper, asserted once rather than 19 times.
func TestRuntime_hasItsMapper(t *testing.T) {
	if mappers := mappersOf(&application.Runtime{}); len(mappers) == 0 {
		t.Error("application.Runtime has no From* mapper, so no runtime maps the application view")
	}
}

// Every mapper takes a context and a *diag.Diagnostics, whatever payload sits
// between them. That uniformity is what lets a reader of 45 files pattern-match
// instead of reading.
func TestResources_mappersFollowTheConvention(t *testing.T) {
	states := map[string]any{"application.Runtime": &application.Runtime{}}
	for typeName, state := range mappedState {
		states[typeName] = state
	}

	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	diagsType := reflect.TypeOf((*diag.Diagnostics)(nil))

	for typeName, state := range states {
		for _, name := range mappersOf(state) {
			method, _ := reflect.TypeOf(state).MethodByName(name)
			signature := method.Type

			// in: receiver, ctx, payload, diags
			if signature.NumIn() != 4 {
				t.Errorf("%s: %s takes %d arguments, want (ctx, payload, diags)",
					typeName, name, signature.NumIn()-1)
				continue
			}
			if signature.In(1) != ctxType {
				t.Errorf("%s: %s takes %s first, want context.Context", typeName, name, signature.In(1))
			}
			if signature.In(3) != diagsType {
				t.Errorf("%s: %s takes %s last, want *diag.Diagnostics", typeName, name, signature.In(3))
			}
		}
	}
}

// mappersOf lists the From* methods declared on a state struct, sorted so a
// failure reads the same way twice.
func mappersOf(state any) []string {
	t := reflect.TypeOf(state)

	var names []string
	for i := range t.NumMethod() {
		if name := t.Method(i).Name; strings.HasPrefix(name, "From") {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
}
