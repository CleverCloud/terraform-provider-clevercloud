package materiakv

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

// mapping.go holds every API-to-state mapper for clevercloud_materia_kv.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view.
//
// The Materia payload carries neither name nor region — they live on the add-on
// itself — so this is the only source of name, region and creation_date. Without
// it an import leaves name null and the required attribute fails the plan (#447).
func (kv *MateriaKV) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *MateriaKV {
	if kv == nil || addon == nil {
		return kv
	}

	kv.Name = pkg.FromStr(addon.Name)
	kv.Region = pkg.FromStr(addon.Region)
	kv.CreationDate = pkg.FromI(addon.CreationDate)

	return kv
}

// FromMateriaDB maps the product view: the connection details.
func (kv *MateriaKV) FromMateriaDB(ctx context.Context, api *models.MateriaDB, diags *diag.Diagnostics) *MateriaKV {
	if kv == nil || api == nil {
		return kv
	}

	kv.Host = pkg.FromStr(api.Host)
	kv.Port = pkg.FromI(int64(api.Port))
	kv.Token = pkg.FromStr(api.Token)

	return kv
}
