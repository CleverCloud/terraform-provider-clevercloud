package pkg

import "regexp"

var (
	AppRegExp = regexp.
			MustCompile(`^app_[0-9a-f]{8}\-[0-9a-f]{4}\-[0-9a-f]{4}\-[0-9a-f]{4}\-[0-9a-f]{12}$`)

	AddonRegExp = regexp.
			MustCompile(`^addon_[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}$`)

	// ServiceRegExp matches the real ID of an add-on: its provider's prefix, then
	// a UUID, or a ULID for Materia KV. The UUID is accepted with and without its
	// hyphens: providers write it hyphenated, but Pulsar add-ons created between
	// mid-2026 and November 2026 got theirs without, and keep it
	ServiceRegExp = regexp.
			MustCompile(`^(?:(?:bucket|cellar|config|elasticsearch|keycloak|matomo|metabase|mongodb|mysql|ng|otoroshi|postgresql|pulsar|redis)_(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}|[0-9a-fA-F]{32})|kv_[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26})$`)

	VhostCleverAppsRegExp = regexp.MustCompile(`^app-.*\.cleverapps\.io$`)

	VhostValidRegExp = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

	RFC3339Regexp = regexp.
			MustCompile(`^((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)$`)
)
