// +build dev

package kumacni

import (
	"net/http"
	"path/filepath"
	"runtime"
)

var Templates http.FileSystem = http.Dir(TemplatesDir(kumactlSrcDir()))

func kumactlSrcDir() string {
	_, thisFile, _, _ := runtime.Caller(1)

	thisDir := filepath.Dir(thisFile)

	return filepath.Join(thisDir, "..", "..", "..", "..")
}
