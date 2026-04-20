package registry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/actions"
	"go.clever-cloud.com/terraform-provider/pkg/datasources/defaultloadbalancer"
	"go.clever-cloud.com/terraform-provider/pkg/datasources/postgresqlbackup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/docker"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/dotnet"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/frankenphp"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/golang"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/java"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/linux"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/nodejs"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/php"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/play2"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/python"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/ruby"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/rust"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/scala"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/static"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/v"
	"go.clever-cloud.com/terraform-provider/pkg/resources/configprovider"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/cellar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/cellar/bucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/elasticsearch"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/fsbucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/materiakv"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mongodb"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mysql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/postgresql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/pulsar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/redis"
	"go.clever-cloud.com/terraform-provider/pkg/resources/drain"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes/nodegroup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/networkgroup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/keycloak"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/matomo"

	"go.clever-cloud.com/terraform-provider/pkg/resources/oauth_consumer"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/metabase"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/otoroshi"
)

var Datasources = []func() datasource.DataSource{
	defaultloadbalancer.NewDataSourceDefaultLoadBalancer,
	postgresqlbackup.NewDataSourcePostgreSQLBackup,
}

var Resources = []func() resource.Resource{
	addon.NewResourceAddon,
	bucket.NewResourceCellarBucket,
	cellar.NewResourceCellar,
	fsbucket.NewResourceFSBucket,
	java.NewResourceJava("war"),
	java.NewResourceJava("jar"),
	linux.NewResourceLinux,
	materiakv.NewResourceMateriaKV,
	metabase.NewResourceMetabase,
	mongodb.NewResourceMongoDB,
	mysql.NewResourceMySQL,
	nodejs.NewResourceNodeJS,
	otoroshi.NewResourceOtoroshi,
	php.NewResourcePHP,
	postgresql.NewResourcePostgreSQL,
	elasticsearch.NewResourceElasticsearch,
	python.NewResourcePython,
	ruby.NewResourceRuby,
	scala.NewResourceScala(),
	static.NewResourceStatic(),
	docker.NewResourceDocker,
	drain.NewDatadogDrain,
	drain.NewNewRelicDrain,
	drain.NewElasticsearchDrain,
	drain.NewSyslogUDPDrain,
	drain.NewSyslogTCPDrain,
	drain.NewHTTPDrain,
	drain.NewOVHDrain,
	keycloak.NewResourceKeycloak,
	kubernetes.NewResourceKubernetes,
	nodegroup.NewResourceKubernetesNodegroup,
	redis.NewResourceRedis,
	golang.NewResourceGo,
	frankenphp.NewResourceFrankenPHP,
	play2.NewResourcePlay2(),
	pulsar.NewResourcePulsar,
	rust.NewResourceRust,
	networkgroup.NewResourceNetworkgroup,
	v.NewResourceV,
	matomo.NewResourceMatomo,
	configprovider.NewResourceConfigProvider,
	dotnet.NewResourceDotnet,
	oauth_consumer.NewResourceOAuthConsumer,
}

var Actions = []func() action.Action{
	actions.RebootApplication,
	actions.ExecuteDatabaseSQL,
	actions.FSBucketUpload,
}

var Functions = []func() function.Function{
	NewEchoFunction,
}

// With the function.Function implementation
func NewEchoFunction() function.Function {
	return &EchoFunction{}
}

var _ function.Function = &EchoFunction{}

type EchoFunction struct{}

func (f *EchoFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "nullifempty"
}

func (f *EchoFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "return null if empty string",
		Description: "return null if empty string",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:               "input",
				Description:        "Value to echo",
				AllowNullValue:     true,
				AllowUnknownValues: true,
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *EchoFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input string

	// Read Terraform argument data into the variable
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))

	// Set the result to the same data
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))
}
