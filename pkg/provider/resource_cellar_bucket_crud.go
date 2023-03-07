package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	minio "github.com/minio/minio-go/v7"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceCellarBucket) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "ResourceCellarBucket.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*Provider)
	if ok {
		r.cc = provider.cc
		r.org = provider.Organisation
	}
}

// Create a new resource
func (r *ResourceCellarBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	bucket := CellarBucket{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &bucket)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cellarEnvRes := tmp.GetAddonEnv(ctx, r.cc, r.org, bucket.CellarID.ValueString())
	if cellarEnvRes.HasError() {
		resp.Diagnostics.AddError(fmt.Sprintf("create: failed to get cellar env %s", bucket.CellarID.String()), cellarEnvRes.Error().Error())
		return
	}

	minioClient, err := minioClientFromEnvsFor(*cellarEnvRes.Payload())
	if err != nil {
		resp.Diagnostics.AddError("failed to setup S3 client", err.Error())
		return
	}

	err = minioClient.MakeBucket(ctx, bucket.Name.ValueString(), minio.MakeBucketOptions{})
	if err != nil {
		resp.Diagnostics.AddError("failed to create bucket", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, bucket)...)
}

// Read resource information
func (r *ResourceCellarBucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Cellar READ", map[string]interface{}{"request": req})

	var cellar CellarBucket
	resp.Diagnostics.Append(req.State.Get(ctx, &cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// nothing to update yet

	resp.Diagnostics.Append(resp.State.Set(ctx, cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceCellarBucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// nothing to update yet (rename ?)
}

// Delete resource
func (r *ResourceCellarBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var bucket CellarBucket

	resp.Diagnostics.Append(req.State.Get(ctx, &bucket)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "CELLAR BUCKET DELETE", map[string]interface{}{"bucket": bucket})

	cellarEnvRes := tmp.GetAddonEnv(ctx, r.cc, r.org, bucket.CellarID.ValueString())
	if cellarEnvRes.HasError() {
		resp.Diagnostics.AddError("delete: failed to get cellar env", cellarEnvRes.Error().Error())
		return
	}

	minioClient, err := minioClientFromEnvsFor(*cellarEnvRes.Payload())
	if err != nil {
		resp.Diagnostics.AddError("failed to setup S3 client", err.Error())
		return
	}

	err = minioClient.RemoveBucket(ctx, bucket.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete bucket", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// Import resource
func (r *ResourceCellarBucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
