package bucket

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/minio/minio-go/v7/pkg/s3utils"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

var (
	_ resource.Resource                = &ResourceCellarBucket{}
	_ resource.ResourceWithImportState = &ResourceCellarBucket{}
)

var cellarIDRegExp = regexp.MustCompile(`^cellar_[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}$`)

type ResourceCellarBucket struct {
	helper.Configurer
}

func NewResourceCellarBucket() resource.Resource {
	return &ResourceCellarBucket{}
}

func (r *ResourceCellarBucket) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_cellar_bucket"
}

// ImportState takes "<cellar_id>/<bucket_name>": a bucket name alone does not
// tell which Cellar holds it, and a bucket name cannot contain a slash.
// cellar_id is kept as given, real ID or add-on ID, so it matches whichever
// form the configuration references.
func (r *ResourceCellarBucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cellarID, name, _ := strings.Cut(req.ID, "/")
	if !cellarIDRegExp.MatchString(cellarID) && !pkg.AddonRegExp.MatchString(cellarID) {
		resp.Diagnostics.AddError(
			"invalid import ID",
			fmt.Sprintf("expected \"<cellar_id>/<bucket_name>\" with a cellar_id like cellar_xxx or addon_xxx, got %q", req.ID),
		)
		return
	}
	if err := s3utils.CheckValidBucketName(name); err != nil {
		resp.Diagnostics.AddError(
			"invalid import ID",
			fmt.Sprintf("expected \"<cellar_id>/<bucket_name>\", got %q: %s", req.ID, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cellar_id"), cellarID)...)
}
