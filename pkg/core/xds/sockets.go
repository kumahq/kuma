package xds

import (
	"fmt"
	"path/filepath"
)

// TODO inline these in kuma-dp once backward compatibility is removed. https://github.com/kumahq/kuma/issues/7220

// AccessLogSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func AccessLogSocketName(tmpDir, name, mesh string) string {
	return socketName(filepath.Join(tmpDir, fmt.Sprintf("kuma-al-%s-%s", name, mesh)))
}

// MetricsHijackerSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func MetricsHijackerSocketName(tmpDir, name, mesh string) string {
	return socketName(filepath.Join(tmpDir, fmt.Sprintf("kuma-mh-%s-%s", name, mesh)))
}

// MeshMetricsDynamicConfigurationSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func MeshMetricsDynamicConfigurationSocketName(workdir string) string {
	return socketName(filepath.Join(workdir, "kuma-mesh-metric-config"))
}

func OpenTelemetrySocketName(workdir string, backendName string) string {
	return socketName(filepath.Join(workdir, "kuma-otel-"+backendName))
}

func socketName(s string) string {
	trimLen := len(s)
	if trimLen > 98 {
		trimLen = 98
	}
	return s[:trimLen] + ".sock"
}
