// +build dev

package postgres

import (
	"net/http"
	"path/filepath"
	"runtime"
)

var Schema http.FileSystem = http.Dir(SchemaDir(kumactlSrcDir()))

func kumactlSrcDir() string {
	_, thisFile, _, _ := runtime.Caller(1)

	thisDir := filepath.Dir(thisFile)

	return filepath.Join(thisDir, "..", "..", "..", "..", "..")
}
