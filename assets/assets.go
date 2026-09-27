package assets

import "embed"

// FS embeds all files in the assets directory (agent registry, skill definitions, references, workflows).
//
//go:embed all:*
var FS embed.FS
