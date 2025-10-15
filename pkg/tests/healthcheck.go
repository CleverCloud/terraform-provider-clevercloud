package tests

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func HealthCheck(_ctx context.Context, vhost string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(_ctx, timeout)
	defer cancel()

	tflog.Info(ctx, "test application endpoint", map[string]any{
		"vhost": vhost,
	})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			res, err := http.Get(fmt.Sprintf("https://%s", vhost))
			if err != nil {
				tflog.Info(ctx, err.Error())
				continue
			}

			tflog.Info(ctx, "application response", map[string]any{
				"status": res.StatusCode,
			})
			if res.StatusCode == 200 {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}
