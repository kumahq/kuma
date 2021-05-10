package data

import (
	"embed"
	"io/fs"
)

//go:embed install
var InstallData embed.FS

func InstallLoggingFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/logging")
	if err != nil {
		panic(err)
	}
	return fsys
}

func InstallMetricsFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/metrics")
	if err != nil {
		panic(err)
	}
	return fsys
}

func InstallTracingFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/tracing")
	if err != nil {
		panic(err)
	}
	return fsys
}

func InstallGatewayFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/gateway")
	if err != nil {
		panic(err)
	}
	return fsys
}
