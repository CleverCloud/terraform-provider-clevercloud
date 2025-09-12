package bucket

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	minio "github.com/minio/minio-go/v7"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceCellarBucket) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceCellarBucket.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
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

	minioClient, err := s3.MinioClientFromEnvsFor(*cellarEnvRes.Payload())
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
	tflog.Debug(ctx, "Cellar READ", map[string]any{"request": req})

	var cellar CellarBucket
	resp.Diagnostics.Append(req.State.Get(ctx, &cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// nothing to update yet

	resp.Diagnostics.Append(resp.State.Set(ctx, cellar)...)
}

// Update resource
func (r *ResourceCellarBucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[CellarBucket](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[CellarBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that immutable fields haven't changed
	if plan.Name.ValueString() != state.Name.ValueString() {
		resp.Diagnostics.AddError(
			"Bucket name cannot be changed",
			"Bucket names are immutable and cannot be changed after creation. To use a different name, you must destroy and recreate the bucket.",
		)
		return
	}

	if plan.CellarID.ValueString() != state.CellarID.ValueString() {
		resp.Diagnostics.AddError(
			"Cellar ID cannot be changed",
			"The cellar_id is immutable and cannot be changed after creation. To use a different cellar, you must destroy and recreate the bucket.",
		)
		return
	}

	// No updateable attributes in current schema, just maintain current state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceCellarBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var bucket CellarBucket

	resp.Diagnostics.Append(req.State.Get(ctx, &bucket)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "CELLAR BUCKET DELETE", map[string]any{"bucket": bucket})

	cellarEnvRes := tmp.GetAddonEnv(ctx, r.cc, r.org, bucket.CellarID.ValueString())
	if cellarEnvRes.HasError() {
		resp.Diagnostics.AddError("delete: failed to get cellar env", cellarEnvRes.Error().Error())
		return
	}

	minioClient, err := s3.MinioClientFromEnvsFor(*cellarEnvRes.Payload())
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
