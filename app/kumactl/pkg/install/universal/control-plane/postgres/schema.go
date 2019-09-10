package postgres

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/Kong/kuma/app/kumactl/pkg/install/universal/control-plane/postgres".Schema

import (
	"path/filepath"
)

func SchemaDir(kumactlSrcDir string) string {
	return filepath.Join(kumactlSrcDir, "data", "install", "universal", "control-plane", "postgres")
}
