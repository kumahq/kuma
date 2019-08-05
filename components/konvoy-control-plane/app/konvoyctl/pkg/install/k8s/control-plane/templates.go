// +build dev

package controlplane

import (
	"net/http"
	"path/filepath"
	"runtime"
)

var Templates http.FileSystem = http.Dir(filepath.Join(konvoyctlDir(), "data", "install", "k8s"))

func konvoyctlDir() string {
	_, thisFile, _, _ := runtime.Caller(1)

	thisDir := filepath.Dir(thisFile)

	return filepath.Join(thisDir, "..", "..", "..", "..")
}
