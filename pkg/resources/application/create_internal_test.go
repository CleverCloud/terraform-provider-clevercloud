package application

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.dev/sdk/models"
)

// cancel_on_push reaches the API on creation and on update: both build their
// payload with wannabeApplication.
func TestWannabeApplication_CancelOnPushFollowsTheAttribute(t *testing.T) {
	for name, tc := range map[string]struct {
		attribute types.Bool
		want      bool
	}{
		"enabled":  {types.BoolValue(true), true},
		"disabled": {types.BoolValue(false), false},
		"null":     {types.BoolNull(), false},
	} {
		t.Run(name, func(t *testing.T) {
			payload := wannabeApplication(&Runtime{CancelOnPush: tc.attribute}, &models.AvailableInstanceView{})

			if payload.CancelOnPush == nil || *payload.CancelOnPush != tc.want {
				t.Errorf("cancelOnPush = %v, want %v", payload.CancelOnPush, tc.want)
			}
		})
	}
}
