package registry

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/resources/cellar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/cellar/bucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/docker"
	"go.clever-cloud.com/terraform-provider/pkg/resources/dotnet"
	"go.clever-cloud.com/terraform-provider/pkg/resources/drain"
	"go.clever-cloud.com/terraform-provider/pkg/resources/frankenphp"
	"go.clever-cloud.com/terraform-provider/pkg/resources/fsbucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/golang"
	"go.clever-cloud.com/terraform-provider/pkg/resources/java"
	"go.clever-cloud.com/terraform-provider/pkg/resources/keycloak"
	"go.clever-cloud.com/terraform-provider/pkg/resources/materiakv"
	"go.clever-cloud.com/terraform-provider/pkg/resources/matomo"
	"go.clever-cloud.com/terraform-provider/pkg/resources/metabase"
	"go.clever-cloud.com/terraform-provider/pkg/resources/mongodb"
	"go.clever-cloud.com/terraform-provider/pkg/resources/mysql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/networkgroup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/nodejs"
	"go.clever-cloud.com/terraform-provider/pkg/resources/otoroshi"
	"go.clever-cloud.com/terraform-provider/pkg/resources/php"
	"go.clever-cloud.com/terraform-provider/pkg/resources/play2"
	"go.clever-cloud.com/terraform-provider/pkg/resources/postgresql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/pulsar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/python"
	"go.clever-cloud.com/terraform-provider/pkg/resources/redis"
	"go.clever-cloud.com/terraform-provider/pkg/resources/ruby"
	"go.clever-cloud.com/terraform-provider/pkg/resources/rust"
	"go.clever-cloud.com/terraform-provider/pkg/resources/scala"
	"go.clever-cloud.com/terraform-provider/pkg/resources/static"
	"go.clever-cloud.com/terraform-provider/pkg/resources/v"
)

var Datasources = []func() datasource.DataSource{}

var Resources = []func() resource.Resource{
	addon.NewResourceAddon,
	bucket.NewResourceCellarBucket,
	cellar.NewResourceCellar,
	fsbucket.NewResourceFSBucket,
	java.NewResourceJava("war"),
	materiakv.NewResourceMateriaKV,
	metabase.NewResourceMetabase,
	mongodb.NewResourceMongoDB,
	mysql.NewResourceMySQL,
	nodejs.NewResourceNodeJS,
	otoroshi.NewResourceOtoroshi,
	php.NewResourcePHP,
	postgresql.NewResourcePostgreSQL,
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
	redis.NewResourceRedis,
	golang.NewResourceGo,
	frankenphp.NewResourceFrankenPHP,
	play2.NewResourcePlay2(),
	pulsar.NewResourcePulsar,
	rust.NewResourceRust,
	networkgroup.NewResourceNetworkgroup,
	dotnet.NewResourceDotnet,
	v.NewResourceV,
	matomo.NewResourceMatomo,
}
