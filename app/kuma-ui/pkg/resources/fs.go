package resources

import (
	"embed"
	"io/fs"
)

//go:embed data
var GuiData embed.FS

func GuiFS() fs.FS {
	fsys, err := fs.Sub(GuiData, "data")
	if err != nil {
		panic(err)
	}
	return fsys
}
