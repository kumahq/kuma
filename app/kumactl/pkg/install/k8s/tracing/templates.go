package tracing

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/tracing".Templates

import (
	"path/filepath"
)

func TemplatesDir(kumactlSrcDir string) string {
	return filepath.Join(kumactlSrcDir, "data", "install", "k8s", "tracing")
}
