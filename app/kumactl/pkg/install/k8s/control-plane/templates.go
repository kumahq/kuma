package controlplane

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/control-plane".HelmTemplates

import (
	"path/filepath"
)

func HelmTemplatesDir(rootSrcDir string) string {
	return filepath.Join(rootSrcDir, "deployments", "charts", "kuma")
}
