package migrations

import (
	"embed"
	"io/fs"

	"github.com/gowool/pages/internal"
)

//go:embed *
var FS embed.FS

var (
	PgFS     = internal.Must(fs.Sub(FS, "pg"))
	SqliteFS = internal.Must(fs.Sub(FS, "sqlite"))
)
