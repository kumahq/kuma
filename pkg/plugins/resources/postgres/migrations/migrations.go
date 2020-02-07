package migrations

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/Kong/kuma/pkg/plugins/resources/postgres/migrations".Migrations

import (
	"path/filepath"
	"runtime"
)

func MigrationsDir() string {
	_, thisFile, _, _ := runtime.Caller(1)
	thisDir := filepath.Dir(thisFile)

	return filepath.Join(thisDir, "data")
}
