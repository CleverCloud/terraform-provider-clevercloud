package helper

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPlanFrom(t *testing.T) {
	ctx := context.Background()

	type MyPlan struct {
		Region string `tfsdk:"region"`
	}

	schem := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{Required: true, MarkdownDescription: "Region"},
		},
	}
	plan := tfsdk.Plan{
		Schema: schem,
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"region": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"region": tftypes.NewValue(tftypes.String, "par"),
		}),
	}
	diags := &diag.Diagnostics{}

	myPlan := PlanFrom[MyPlan](ctx, plan, diags)
	if diags.HasError() {
		t.Errorf("PlanFrom() expect not errors, got %v", diags)
	}

	if !reflect.DeepEqual(myPlan.Region, "par") {
		t.Errorf("PlanFrom() = %v, want %v", myPlan.Region, "par")
	}

}

func TestPlanFromWithError(t *testing.T) {
	ctx := context.Background()

	type MyPlan struct {
		Region string `tfsdk:"region"`
	}

	schem := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.Int64Attribute{Required: true, MarkdownDescription: "Region"},
		},
	}
	plan := tfsdk.Plan{
		Schema: schem,
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"region": tftypes.Number,
			},
		}, map[string]tftypes.Value{
			"region": tftypes.NewValue(tftypes.Number, 10),
		}),
	}
	diags := &diag.Diagnostics{}

	_ = PlanFrom[MyPlan](ctx, plan, diags)
	if !diags.HasError() {
		t.Errorf("PlanFrom() expect errors, got %v", diags)
	}
}
