// Package db embeds the SQLite schema so the binary is self-contained —
// schema initialization must not depend on the process working directory.
package db

import _ "embed"

//go:embed schema.sql
var Schema string
