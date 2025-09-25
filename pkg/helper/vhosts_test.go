package helper_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

func TestVHostsFromAPIHosts(t *testing.T) {
	objSchema := types.ObjectType{AttrTypes: map[string]attr.Type{"fqdn": types.StringType, "path_begin": types.StringType}}

	tests := []struct {
		name string
		raw  []string
		want types.Set
	}{{
		name: "main",
		raw:  []string{"foo.com", "bar.com/test", "bizz.com/"},
		want: types.SetValueMust(objSchema,
			[]attr.Value{
				types.ObjectValueMust(map[string]attr.Type{"fqdn": types.StringType, "path_begin": types.StringType}, map[string]attr.Value{
					"fqdn":       types.StringValue("foo.com"),
					"path_begin": types.StringValue("/"),
				}),
				types.ObjectValueMust(map[string]attr.Type{"fqdn": types.StringType, "path_begin": types.StringType}, map[string]attr.Value{
					"fqdn":       types.StringValue("bar.com"),
					"path_begin": types.StringValue("/test"),
				}),
				types.ObjectValueMust(map[string]attr.Type{"fqdn": types.StringType, "path_begin": types.StringType}, map[string]attr.Value{
					"fqdn":       types.StringValue("bizz.com"),
					"path_begin": types.StringValue("/"),
				}),
			}),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := diag.Diagnostics{}
			got := helper.VHostsFromAPIHosts(context.Background(), tt.raw, types.SetNull(objSchema), &d)
			if d.HasError() {
				t.Errorf("No diag expected, VHostsFromAPIHosts() = %v", d)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VHostsFromAPIHosts() = %v, want %v", got, tt.want)
			}
		})
	}
}
