package xds

import (
	"fmt"
	"path/filepath"
)

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

// ReadinessReporterSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func ReadinessReporterSocketName(workdir string) string {
	return socketName(filepath.Join(workdir, "kuma-readiness-reporter"))
}

// AdminSocketName generates a socket path for the Envoy admin API that will fit the Unix socket path limitation of 104 chars
func AdminSocketName(workdir string) string {
	return socketName(filepath.Join(workdir, "kuma-envoy-admin"))
}

func socketName(s string) string {
	trimLen := len(s)
	if trimLen > 98 {
		trimLen = 98
	}
	return s[:trimLen] + ".sock"
}
