package version

import (
	"runtime/debug"
)

var (
	// Version is populated via ldflags at build time (e.g., -X github.com/koharness/koharness/pkg/version.Version=v0.1.0).
	Version = ""
	// Commit is populated via ldflags at build time.
	Commit = ""
	// Date is populated via ldflags at build time.
	Date = ""
)

// Info holds version metadata.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Get resolves version info from ldflags, debug.ReadBuildInfo(), or fallback commit ID.
func Get() Info {
	v := Version
	c := Commit
	d := Date

	if info, ok := debug.ReadBuildInfo(); ok {
		if c == "" {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					c = setting.Value
					break
				}
			}
		}
		if d == "" {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.time" {
					d = setting.Value
					break
				}
			}
		}
		if v == "" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
	}

	shortCommit := c
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	if v == "" {
		if shortCommit != "" {
			v = "dev-" + shortCommit
		} else {
			v = "dev"
		}
	}

	return Info{
		Version: v,
		Commit:  shortCommit,
		Date:    d,
	}
}

// GetVersion returns the resolved version string.
func GetVersion() string {
	return Get().Version
}
