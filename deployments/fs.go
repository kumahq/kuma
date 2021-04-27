package deployments

import (
	"embed"
	"io/fs"
)

// By default, go embed does not embed files that starts with `.` or `_` that's why we need to add _helpers.tpl explicitly

//go:embed charts/kuma/templates/* charts/kuma/crds/* charts/kuma/*.yaml
var ChartsData embed.FS

func KumaChartFS() fs.FS {
	fsys, err := fs.Sub(ChartsData, "charts/kuma")
	if err != nil {
		panic(err)
	}
	return fsys
}
