package buildinfo

import (
	"runtime/debug"
	"testing"
)

func info(settings ...debug.BuildSetting) *debug.BuildInfo {
	return &debug.BuildInfo{Settings: settings}
}

func TestResolveCommit(t *testing.T) {
	rev := debug.BuildSetting{Key: "vcs.revision", Value: "deadbeef"}
	dirty := debug.BuildSetting{Key: "vcs.modified", Value: "true"}
	clean := debug.BuildSetting{Key: "vcs.modified", Value: "false"}

	tests := []struct {
		name string
		env  string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{"env var wins over build info", "cafe42", info(rev), true, "cafe42"},
		{"revision from build info", "", info(rev, clean), true, "deadbeef"},
		{"dirty working tree is flagged", "", info(rev, dirty), true, "deadbeef-dirty"},
		{"no revision in build info", "", info(), true, "unknown"},
		{"no build info at all", "", nil, false, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCommit(tt.env, tt.info, tt.ok)
			if got != tt.want {
				t.Errorf("resolveCommit() = %q, want %q", got, tt.want)
			}
		})
	}
}
