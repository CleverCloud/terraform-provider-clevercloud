package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Test if right environment variables are set
func TestNodeJS_GetEnv(t *testing.T) {
	type fields struct {
		ID               types.String
		Name             types.String
		Description      types.String
		MinInstanceCount types.Int64
		MaxInstanceCount types.Int64
		SmallestFlavor   types.String
		BiggestFlavor    types.String
		BuildFlavor      types.String
		Region           types.String
		StickySessions   types.Bool
		RedirectHTTPS    types.Bool
		Commit           types.String
		VHost            types.String
		AdditionalVHosts types.List
		DeployURL        types.String
		AppFolder        types.String
		DevDependencies  types.Bool
		StartScript      types.String
		PackageManager   types.String
		Registry         types.String
		RegistryToken    types.String
		Environment      types.Map
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{{
		name: "main",
		fields: fields{
			Environment: types.Map{
				Elems: map[string]attr.Value{
					"MY_ENV": fromStr("A"),
				},
			},
			PackageManager: fromStr("yarn"),
			AppFolder:      fromStr("/home"),
		},
		want: map[string]string{
			"MY_ENV":             "A",
			"CC_NODE_BUILD_TOOL": "yarn",
			"APP_FOLDER":         "/home",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := NodeJS{
				ID:               tt.fields.ID,
				Name:             tt.fields.Name,
				Description:      tt.fields.Description,
				MinInstanceCount: tt.fields.MinInstanceCount,
				MaxInstanceCount: tt.fields.MaxInstanceCount,
				SmallestFlavor:   tt.fields.SmallestFlavor,
				BiggestFlavor:    tt.fields.BiggestFlavor,
				BuildFlavor:      tt.fields.BuildFlavor,
				Region:           tt.fields.Region,
				StickySessions:   tt.fields.StickySessions,
				RedirectHTTPS:    tt.fields.RedirectHTTPS,
				Commit:           tt.fields.Commit,
				VHost:            tt.fields.VHost,
				AdditionalVHosts: tt.fields.AdditionalVHosts,
				DeployURL:        tt.fields.DeployURL,
				AppFolder:        tt.fields.AppFolder,
				Environment:      tt.fields.Environment,
				DevDependencies:  tt.fields.DevDependencies,
				StartScript:      tt.fields.StartScript,
				PackageManager:   tt.fields.PackageManager,
				Registry:         tt.fields.Registry,
				RegistryToken:    tt.fields.RegistryToken,
			}
			got, diag := plan.GetEnv(context.Background())
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("NodeJS.getEnv() for '%s' got = %v, want %v", k, got[k], v)
				}
			}
			if diag.HasError() {
				t.Errorf("NodeJS.getEnv() error: %v", diag)
			}
		})
	}
}
