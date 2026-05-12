package elasticsearch_cluster

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"go.clever-cloud.com/terraform-provider/pkg"
)

type ElasticsearchCluster struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	NetworkGroupID types.String `tfsdk:"networkgroup_id"`
	Version        types.Object `tfsdk:"version"`
	NodeCount      types.Int64  `tfsdk:"node_count"`
	CPUCount       types.Int64  `tfsdk:"cpu_count"`
	MemorySize     types.Int64  `tfsdk:"memory_size"`
	DiskSize       types.Int64  `tfsdk:"disk_size"`
	Endpoint       types.String `tfsdk:"endpoint"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
}

type Version struct {
	Major types.Int64 `tfsdk:"major"`
	Minor types.Int64 `tfsdk:"minor"`
	Patch types.Int64 `tfsdk:"patch"`
}

var versionAttrTypes = map[string]attr.Type{
	"major": types.Int64Type,
	"minor": types.Int64Type,
	"patch": types.Int64Type,
}

//go:embed doc.md
var resourceDoc string

func (r ResourceElasticsearchCluster) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	atLeastOne := []validator.Int64{pkg.NewInt64AtLeastValidator(1)}

	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceDoc,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the Elasticsearch cluster",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Elasticsearch cluster",
			},
			"networkgroup_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network group ID. If not provided, the API will assign one automatically",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"version": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Elasticsearch version. Fields left unset will be chosen by the API",
				Validators:          []validator.Object{validateVersionSemver},
				Attributes: map[string]schema.Attribute{
					"major": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Major version number",
					},
					"minor": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Minor version number",
					},
					"patch": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Patch version number",
					},
				},
			},
			"node_count": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             pkg.StaticInt64(1),
				MarkdownDescription: "Number of nodes in the cluster",
				Validators:          atLeastOne,
			},
			"cpu_count": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             pkg.StaticInt64(4),
				MarkdownDescription: "Number of CPUs per node",
				Validators:          atLeastOne,
			},
			"memory_size": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             pkg.StaticInt64(8192),
				MarkdownDescription: "Memory per node in MB",
				Validators:          atLeastOne,
			},
			"disk_size": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             pkg.StaticInt64(25600),
				MarkdownDescription: "Disk size per node in MB",
				Validators:          atLeastOne,
			},
			"endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Endpoint to connect to the Elasticsearch cluster",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Username for Elasticsearch authentication",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Password for Elasticsearch authentication",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

var validateVersionSemver = pkg.NewObjectValidator(
	"version fields must follow semver hierarchy: major must be set before minor, minor before patch",
	func(ctx context.Context, req validator.ObjectRequest, res *validator.ObjectResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		var v Version
		diags := req.ConfigValue.As(ctx, &v, basetypes.ObjectAsOptions{})
		res.Diagnostics.Append(diags...)
		if res.Diagnostics.HasError() {
			return
		}

		hasMajor := !v.Major.IsNull() && !v.Major.IsUnknown()
		hasMinor := !v.Minor.IsNull() && !v.Minor.IsUnknown()
		hasPatch := !v.Patch.IsNull() && !v.Patch.IsUnknown()

		if hasMinor && !hasMajor {
			res.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid version specification",
				"Cannot set minor version without major version",
			)
		}

		if hasPatch && !hasMinor {
			res.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid version specification",
				"Cannot set patch version without minor version",
			)
		}
	},
)

