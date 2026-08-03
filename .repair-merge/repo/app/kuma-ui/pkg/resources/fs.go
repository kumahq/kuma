package resources

import (
	"embed"
	"io/fs"
)

// By default, go embed does not embed files that starts with `_` that's why we need to use *

//go:embed data/* data/assets/*
var GuiData embed.FS

var GuiFS = func() fs.FS {
	fsys, err := fs.Sub(GuiData, "data")
	if err != nil {
		panic(err)
	}
	return fsys
}
