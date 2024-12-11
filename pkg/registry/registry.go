package registry

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/resources/cellar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/cellar/bucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/docker"
	"go.clever-cloud.com/terraform-provider/pkg/resources/java"
	"go.clever-cloud.com/terraform-provider/pkg/resources/keycloak"
	"go.clever-cloud.com/terraform-provider/pkg/resources/materiakv"
	"go.clever-cloud.com/terraform-provider/pkg/resources/metabase"
	"go.clever-cloud.com/terraform-provider/pkg/resources/mongodb"
	"go.clever-cloud.com/terraform-provider/pkg/resources/nodejs"
	"go.clever-cloud.com/terraform-provider/pkg/resources/php"
	"go.clever-cloud.com/terraform-provider/pkg/resources/postgresql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/python"
	"go.clever-cloud.com/terraform-provider/pkg/resources/scala"
	"go.clever-cloud.com/terraform-provider/pkg/resources/static"
)

var Datasources = []func() datasource.DataSource{}

var Resources = []func() resource.Resource{
	addon.NewResourceAddon,
	bucket.NewResourceCellarBucket,
	cellar.NewResourceCellar,
	java.NewResourceJava("war"),
	materiakv.NewResourceMateriaKV,
	metabase.NewResourceMetabase,
	mongodb.NewResourceMongoDB,
	nodejs.NewResourceNodeJS,
	php.NewResourcePHP,
	postgresql.NewResourcePostgreSQL,
	python.NewResourcePython,
	scala.NewResourceScala(),
	static.NewResourceStatic(),
	docker.NewResourceDocker,
	keycloak.NewResourceKeycloak,
}
