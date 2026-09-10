package data

import (
	"embed"
	"io/fs"
)

//go:embed install
var InstallData embed.FS

func InstallDemoFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/demo")
	if err != nil {
		panic(err)
	}
	return fsys
}
