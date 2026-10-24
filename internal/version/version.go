// Package version holds the build-time identity of the rankcore binary.
package version

// Version is the semantic version of the binary. Release builds override it with
// -ldflags "-X github.com/zimkk/rankcore/internal/version.Version=v1.2.3".
var Version = "1.1.0"

// BuildTime is the RFC 3339 build timestamp, injected by release builds.
var BuildTime = ""

const (
	// RulesetVersion identifies the deterministic rule catalog.
	RulesetVersion = "2026.10"

	// SchemaVersion is the version of the audit.json / verification.json contract.
	SchemaVersion = 1
)

// UserAgent is the transparent crawler identity used for remote crawls.
func UserAgent() string {
	return "RankCoreBot/" + Version + " (+https://github.com/zimkk/rankcore)"
}
