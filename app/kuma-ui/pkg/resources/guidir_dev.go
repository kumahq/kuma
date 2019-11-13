// +build dev

package resources

import (
	"net/http"
	"path/filepath"
	"runtime"
)

var GuiDir http.FileSystem = http.Dir(guiSrcDir())

func guiSrcDir() string {
	_, thisFile, _, _ := runtime.Caller(1)

	thisDir := filepath.Dir(thisFile)

	return filepath.Join(thisDir, "..", "..", "data", "resources")
}
