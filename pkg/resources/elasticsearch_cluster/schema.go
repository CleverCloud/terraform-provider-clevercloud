package elasticsearch_cluster

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
	Plan           types.String `tfsdk:"plan"`
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
	// The API refuses clusters under 3 nodes
	atLeastThree := []validator.Int64{pkg.NewInt64AtLeastValidator(3)}
	// The cluster API has no update endpoint: every user-facing attribute is
	// immutable and a change recreates the cluster.
	replaceStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	versionPart := []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}

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
				PlanModifiers:       replaceStr,
			},
			"networkgroup_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network group the cluster is reachable from. If not provided, the API creates a dedicated one",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"version": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Elasticsearch version. Fields left unset will be chosen by the API",
				Validators:          []validator.Object{validateVersionSemver},
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown(), objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"major": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Major version number",
						PlanModifiers:       versionPart,
					},
					"minor": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Minor version number",
						PlanModifiers:       versionPart,
					},
					"patch": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Patch version number",
						PlanModifiers:       versionPart,
					},
				},
			},
			"node_count": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             pkg.StaticInt64(3),
				MarkdownDescription: "Number of nodes in the cluster (at least 3)",
				Validators:          atLeastThree,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"plan": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Node plan defining CPU, memory and disk per node (e.g. M, L, XL)",
				PlanModifiers:       replaceStr,
			},
			"endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Host of the Elasticsearch cluster, only reachable from inside its network group (`networkgroup_id`)",
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
