package pkg

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type int64Default struct {
	val int64
}

func StaticInt64(val int64) defaults.Int64 {
	return &int64Default{val: val}
}

func (d *int64Default) Description(context.Context) string {
	return fmt.Sprintf("defaults to %d", d.val)
}

func (d *int64Default) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *int64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.val)
}
