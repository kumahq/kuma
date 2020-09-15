// +build dev

package controlplane

import (
	"net/http"
	"path/filepath"
	"runtime"
)

var HelmTemplates http.FileSystem = http.Dir(HelmTemplatesDir(rootSrcDir()))

func rootSrcDir() string {
	_, thisFile, _, _ := runtime.Caller(1)

	thisDir := filepath.Dir(thisFile)

	return filepath.Join(thisDir, "..", "..", "..", "..", "..", "..")
}
