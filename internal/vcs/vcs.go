package vcs

import (
	"fmt"
	"runtime/debug"
)

// Version returns a version number of the form
//
//	`<time>-<revision-number>[-dirty]`
//
// where <revision-number> is the current version control revision number,
// <time> is the time the revision was made, and and [-dirty] is included if
// the code in the binary has modifications that aren't included in the
// revision.
//
// This information is obtained via debug.ReadBuildInfo(), which provides at
// runtime the same information that is provided by `go version -m <binary>`.
func Version() string {
	var (
		revision string
		time     string
		modified bool
	)

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.time":
				time = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}

	if modified {
		return fmt.Sprintf("%s-%s-dirty", time, revision)
	}

	return fmt.Sprintf("%s-%s", time, revision)
}
