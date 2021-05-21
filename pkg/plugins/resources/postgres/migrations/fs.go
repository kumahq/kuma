package migrations

import (
	"embed"
	"io/fs"
)

//go:embed data
var Data embed.FS

func MigrationFS() fs.FS {
	fsys, err := fs.Sub(Data, "data")
	if err != nil {
		panic(err)
	}
	return fsys
}
