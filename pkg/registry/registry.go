package registry

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources/cellar"
	"go.clever-cloud.com/terraform-provider/pkg/resources/cellar/bucket"
	"go.clever-cloud.com/terraform-provider/pkg/resources/java"
	"go.clever-cloud.com/terraform-provider/pkg/resources/nodejs"
	"go.clever-cloud.com/terraform-provider/pkg/resources/php"
	"go.clever-cloud.com/terraform-provider/pkg/resources/postgresql"
	"go.clever-cloud.com/terraform-provider/pkg/resources/scala"
)

var Datasources = []func() datasource.DataSource{}

var Resources = []func() resource.Resource{
	cellar.NewResourceCellar,
	bucket.NewResourceCellarBucket,
	nodejs.NewResourceNodeJS,
	php.NewResourcePHP,
	postgresql.NewResourcePostgreSQL,
	java.NewResourceJava("war"),
	scala.NewResourceScala(),
}
