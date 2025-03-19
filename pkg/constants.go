package pkg

import (
	"runtime/debug"
)

var (
	Commit    = "0000000000000000000000000000000000000000"
	SentryDSN = "https://9f1c17cd85db40f5a1991aefcc182944@glitchtip.corp.clever-cloud.com/35"
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				Commit = setting.Value
			}
		}
	}
}
