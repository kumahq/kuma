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

func InstallDeprecatedLoggingFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s-deprecated/logging")
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

func InstallDeprecatedMetricsFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s-deprecated/metrics")
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

func InstallDeprecatedTracingFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s-deprecated/tracing")
	if err != nil {
		panic(err)
	}
	return fsys
}

func InstallDemoFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/demo")
	if err != nil {
		panic(err)
	}
	return fsys
}

func InstallGatewayKongFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/gateway-kong")
	if err != nil {
		panic(err)
	}
	return fsys
}

func InstallGatewayKongEnterpriseFS() fs.FS {
	fsys, err := fs.Sub(InstallData, "install/k8s/gateway-kong-enterprise")
	if err != nil {
		panic(err)
	}
	return fsys
}
