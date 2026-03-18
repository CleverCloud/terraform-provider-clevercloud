package materiakv

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
)

type MateriaKV struct {
	addon.CommonAttributes
	Host  types.String `tfsdk:"host"`
	Port  types.Int64  `tfsdk:"port"`
	Token types.String `tfsdk:"token"`
}

func (kv *MateriaKV) GetCommonPtr() *addon.CommonAttributes {
	return &kv.CommonAttributes
}

func (kv *MateriaKV) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (kv *MateriaKV) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	s := sdk.NewSDK(sdk.WithClient(cc))
	kvInfoRes := s.V4().Materia().
		Organisations().Ownerid(org).Materia().
		Databases().Resourceid(kv.ID.ValueString()).Getmateriakvv4(ctx)
	if kvInfoRes.HasError() {
		diags.AddError("failed to get materia kv connection infos", kvInfoRes.Error().Error())
		return
	}

	kvInfo := kvInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", kvInfo),
	})

	if kvInfo.Status == "TO_DELETE" {
		diags.AddError("addon is being deleted", "MateriaKV addon is marked for deletion")
		return
	}

	kv.Host = pkg.FromStr(kvInfo.Host)
	kv.Port = pkg.FromI(int64(kvInfo.Port))
	kv.Token = pkg.FromStr(kvInfo.Token)
}

func (kv *MateriaKV) SetDefaults() {}

//go:embed doc.md
var resourceMateriaKVDoc string

func (r ResourceMateriaKV) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMateriaKVDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// Single-plan addon: plan is computed, not user-specified
			"plan":  schema.StringAttribute{Computed: true},
			"host":  schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"token": schema.StringAttribute{Computed: true, MarkdownDescription: "Token to authenticate", Sensitive: true},
		}),
	}
}
