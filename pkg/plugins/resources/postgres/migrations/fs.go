package migrations

import (
	"embed"
	"io/fs"
)

//go:embed data
var Data embed.FS

var MigrationFS func() fs.FS

func init() {
	MigrationFS = func() fs.FS {
		fsys, err := fs.Sub(Data, "data")
		if err != nil {
			panic(err)
		}
		return fsys
	}
}
