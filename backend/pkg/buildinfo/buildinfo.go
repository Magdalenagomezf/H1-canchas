// Package buildinfo reports which commit the running binary was built from.
package buildinfo

import (
	"os"
	"runtime/debug"
)

// CommitEnvVar lets the deploy pipeline set the commit explicitly. It is
// useful when the host builds without the .git directory, so Go cannot
// embed the revision.
const CommitEnvVar = "APP_COMMIT"

// HostCommitEnvVar is set by Render at build and run time to the commit it
// deployed. The deploy pipeline pins that commit, so it is the one the smoke
// test expects even when the build did not embed the revision.
const HostCommitEnvVar = "RENDER_GIT_COMMIT"

// Commit returns the commit of the running binary: APP_COMMIT if set,
// otherwise RENDER_GIT_COMMIT, otherwise the git revision Go embeds at build
// time ("-dirty" when built with uncommitted changes), otherwise "unknown"
// (for example under go run).
func Commit() string {
	info, ok := debug.ReadBuildInfo()
	return resolveCommit(os.Getenv(CommitEnvVar), os.Getenv(HostCommitEnvVar), info, ok)
}

func resolveCommit(env, host string, info *debug.BuildInfo, ok bool) string {
	if env != "" {
		return env
	}
	if host != "" {
		return host
	}
	if !ok || info == nil {
		return "unknown"
	}
	var revision string
	var modified bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if revision == "" {
		return "unknown"
	}
	if modified {
		return revision + "-dirty"
	}
	return revision
}
