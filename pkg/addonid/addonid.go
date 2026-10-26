// Package addonid converts between the two identifiers of a Clever Cloud add-on:
// its add-on ID (`addon_xxx`), which the v2 API addresses it by, and its real ID
// (`postgresql_xxx`, `mysql_xxx`, ...), which the provider's v4 APIs and the
// Terraform resources use.
package addonid

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/retry"
	"go.clever-cloud.dev/client"
)

// FromRealIDs resolves real IDs (`postgresql_xxx`) to add-on IDs (`addon_xxx`); an
// add-on ID is returned as is.
func FromRealIDs(ctx context.Context, cc *client.Client, organisation string, realID ...string) ([]string, error) {
	addonsRes := pkg.SDK(cc).V2().Organisations().ID(organisation).Addons().Getaddonsbyorgaid(ctx)
	if addonsRes.HasError() {
		return nil, addonsRes.Error()
	}
	addons := *addonsRes.Payload()

	addonIDs := make([]string, len(realID))
	for i, id := range realID {
		if strings.HasPrefix(id, "addon_") {
			addonIDs[i] = id
			continue
		}

		notFound := true

		for _, addon := range addons {
			if pkg.Deref(addon.RealID) == id {
				addonIDs[i] = addon.ID
				notFound = false
			}
		}

		if notFound {
			return nil, fmt.Errorf("addon %s not found", id)
		}
	}

	return addonIDs, nil
}

// FromRealID resolves a real ID (`postgresql_xxx`) to its add-on ID (`addon_xxx`),
// retrying while the add-on is not listed yet; an add-on ID is returned as is.
func FromRealID(ctx context.Context, cc *client.Client, organisation string, realID string) (string, error) {
	if strings.HasPrefix(realID, "addon_") {
		return realID, nil
	}

	// Use retry mechanism with custom config for addon listing
	config := retry.DefaultConfig()
	var lastError error

	// Wrap the operation in a retry mechanism
	for attempt := range config.MaxAttempts {
		addonsRes := pkg.SDK(cc).V2().Organisations().ID(organisation).Addons().Getaddonsbyorgaid(ctx)
		if addonsRes.HasError() {
			lastError = addonsRes.Error()
			tflog.Warn(ctx, fmt.Sprintf(
				"FromRealID: failed to list addons (attempt %d/%d): %v",
				attempt+1,
				config.MaxAttempts,
				lastError,
			))
			continue
		}
		addons := *addonsRes.Payload()

		for _, addon := range addons {
			if pkg.Deref(addon.RealID) == realID {
				return addon.ID, nil
			}
		}

		// Addon not found, retry after a delay
		lastError = fmt.Errorf("addon %s not found in list", realID)
		backoff := min(time.Duration(float64(config.InitialDelay)*math.Pow(config.Multiplier, float64(attempt))), config.MaxDelay)

		tflog.Warn(ctx, fmt.Sprintf(
			"FromRealID: addon %s not found in list (attempt %d/%d), retrying in %v",
			realID,
			attempt+1,
			config.MaxAttempts,
			backoff,
		))

		select {
		case <-ctx.Done():
			tflog.Error(ctx, fmt.Sprintf("FromRealID: context cancelled while retrying to find addon %s", realID))
			return "", fmt.Errorf("context cancelled while retrying to find addon %s: %w", realID, ctx.Err())
		case <-time.After(backoff):
			// Continue to next attempt
		}

	}

	tflog.Error(ctx, fmt.Sprintf(
		"FromRealID: addon %s not found after %d attempts",
		realID,
		config.MaxAttempts,
	))
	return "", fmt.Errorf("addon %s not found after %d attempts: %w", realID, config.MaxAttempts, lastError)
}

// ToRealID resolves an add-on ID (`addon_xxx`) to its real ID (`postgresql_xxx`); a
// real ID is returned as is.
func ToRealID(ctx context.Context, cc *client.Client, organisation string, addonID string) (string, error) {
	if !strings.HasPrefix(addonID, "addon_") {
		return addonID, nil
	}

	addonsRes := pkg.SDK(cc).V2().Organisations().ID(organisation).Addons().Getaddonsbyorgaid(ctx)
	if addonsRes.HasError() {
		return "", addonsRes.Error()
	}
	addons := *addonsRes.Payload()

	for _, addon := range addons {
		if addon.ID == addonID {
			return pkg.Deref(addon.RealID), nil
		}
	}

	return "", fmt.Errorf("addon %s not found", addonID)
}
