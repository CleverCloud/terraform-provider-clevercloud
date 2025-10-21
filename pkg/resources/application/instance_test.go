package application

import (
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func TestAccNodejs_basic(t *testing.T) {
	t.Parallel()

	variants := []tmp.ProductInstance{{
		Version: "20220101",
	}, {
		Version: "20230503",
	}, {
		Version: "20230502",
	}, {
		Version: "201901209",
	}}

	last := lastVariant(variants)

	expected := "20230503"
	if last.Version != expected {
		t.Errorf("expect '%s' as variant, but got '%s'", expected, last.Version)
	}
}
