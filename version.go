package main

import (
	"fmt"
)

// Version information
var (
	// Version is the current version of the application
	// This will be overwritten during build by ldflags
	Version = "dev"

	// GitCommit is the git commit hash
	// This will be overwritten during build by ldflags
	GitCommit = "unknown"

	// BuildTime is the time when the binary was built
	// This will be overwritten during build by ldflags
	BuildTime = "unknown"
)

// VersionInfo returns formatted version information
func VersionInfo() string {
	return fmt.Sprintf("Docker Swarm Exporter\nVersion: %s\nGit commit: %s\nBuild time: %s",
		Version, GitCommit, BuildTime)
}
