package pkg

import "regexp"

var (
	AppRegExp = regexp.MustCompile(
		`^app_[0-9a-f]{8}\-[0-9a-f]{4}\-[0-9a-f]{4}\-[0-9a-f]{4}\-[0-9a-f]{12}$`,
	)

	AddonRegExp = regexp.MustCompile(
		`^addon_[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}$`,
	)

	ServiceRegExp = regexp.MustCompile(
		`^(postgresql|redis|cellar|config|matomo|mysql|pulsar|bucket)_[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}$`,
	)

	VhostCleverAppsRegExp = regexp.MustCompile(`^app-.*\.cleverapps\.io$`)
)
