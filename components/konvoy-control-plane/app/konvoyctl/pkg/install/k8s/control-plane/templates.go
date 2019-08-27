package controlplane

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/install/k8s/control-plane".Templates

import (
	"path/filepath"
)

func TemplatesDir(konvoyctlSrcDir string) string {
	return filepath.Join(konvoyctlSrcDir, "data", "install", "k8s")
}
