package metrics

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/Kong/kuma/app/kumactl/pkg/install/k8s/metrics".Templates

import (
	"path/filepath"
)

func TemplatesDir(kumactlSrcDir string) string {
	return filepath.Join(kumactlSrcDir, "data", "install", "k8s", "metrics")
}
