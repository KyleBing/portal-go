package migrations

import "embed"

// FS embeds incremental schema migrations (NNN_*.sql).
// init.sql is used only by the first-time setup flow and is not applied here.
//
//go:embed *.sql
var FS embed.FS
