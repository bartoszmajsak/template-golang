package version

import "strconv"

var (
	// Version holds a semantic version of the running binary.
	Version = "v0.0.0"
	// Commit holds the commit hash against which the binary build was run.
	Commit string
	// BuildTime holds timestamp when the binary build was run.
	BuildTime string
	// Release indicates if built binary is an official release.
	Release = "false"
)

func Released() bool {
	released, _ := strconv.ParseBool(Release)

	return released
}
