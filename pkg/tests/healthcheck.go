package tests

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func HealthCheck(_ctx context.Context, vhost string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(_ctx, timeout)
	defer cancel()

	fmt.Printf("Test on %s\n", vhost)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			res, err := http.Get(fmt.Sprintf("https://%s", vhost))
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}

			fmt.Printf("RESPONSE %d\n", res.StatusCode)
			if res.StatusCode == 200 {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}
