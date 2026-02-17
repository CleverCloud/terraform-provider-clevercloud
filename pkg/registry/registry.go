package registry

import (
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/fsbucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/materiakv"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mongodb"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/mysql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/postgresql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/pulsar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/redis"
	"go.clever-cloud.com/terraform-provider/pkg/resources/drain"
	"go.clever-cloud.com/terraform-provider/pkg/resources/elasticsearch"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes/nodegroup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/networkgroup"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/keycloak"
	"go.clever-cloud.com/terraform-provider/pkg/resources/software/matomo"

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
}

var Actions = []func() action.Action{
	actions.RebootApplication,
	actions.ExecuteDatabaseSQL,
	actions.FSBucketUpload,
}
