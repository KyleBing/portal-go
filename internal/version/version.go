package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var raw string

// Version is the single source of truth for releases / deploy logs.
var Version = strings.TrimSpace(raw)
