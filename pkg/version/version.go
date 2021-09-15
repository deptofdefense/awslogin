package version

import (
	"fmt"
	"runtime"
)

// current version
const (
	coreVersion = "0.1.0"
	// Values should be: "", alpha, beta, rc1, rcN
	// An empty string is the released core version
	prerelease = ""
)

// Provisioned by ldflags
var commit string

// Core return the core version.
func Core() string {
	return coreVersion
}

// Short return the version with pre-release, if available.
func Short() string {
	v := coreVersion

	if prerelease != "" {
		v += "-" + prerelease
	}

	return v
}

// Full return the full version including pre-release, commit hash, runtime os and arch.
func Full() string {
	if commit != "" && commit[:1] != " " {
		commit = " " + commit
	}

	return fmt.Sprintf("v%s%s %s/%s", Short(), commit, runtime.GOOS, runtime.GOARCH)
}
