package tests

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// HealthCheck polls https://vhost until it answers 200 or timeout expires.
// Each request is bound to the deadline, so a hanging connection can't
// outlive it.
func HealthCheck(_ctx context.Context, vhost string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(_ctx, timeout)
	defer cancel()

	tflog.Info(ctx, "test application endpoint", map[string]any{
		"vhost": vhost,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s", vhost), nil)
		if err != nil {
			return err
		}

		res, err := client.Do(req)
		if err != nil {
			tflog.Info(ctx, err.Error())
		} else {
			_ = res.Body.Close()
			tflog.Info(ctx, "application response", map[string]any{
				"status": res.StatusCode,
			})
			if res.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
