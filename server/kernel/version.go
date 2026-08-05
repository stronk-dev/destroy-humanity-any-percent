package kernel

import "runtime/debug"

// Version is generated from the repository's kernel/VERSION source of truth.
// TestVersionSourceOfTruth is the fail-closed drift gate until code generation
// is wired into the build command.
const Version = "0.3.38"

func VCSRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			return setting.Value
		}
	}
	return "unknown"
}
