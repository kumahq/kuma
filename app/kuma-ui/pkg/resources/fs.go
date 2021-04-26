package resources

import (
	"embed"
	"io/fs"
)

//go:embed data
var GuiData embed.FS

var GuiFS = func() fs.FS {
	fsys, err := fs.Sub(GuiData, "data")
	if err != nil {
		panic(err)
	}
	return fsys
}
