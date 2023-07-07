package xds

import (
	"fmt"
	"path/filepath"
)

// TODO inline these in kuma-dp once backward compatibility is removed.

// AccessLogSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func AccessLogSocketName(tmpDir, name, mesh string) string {
	return socketName(filepath.Join(tmpDir, fmt.Sprintf("kuma-al-%s-%s", name, mesh)))
}

// MetricsHijackerSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func MetricsHijackerSocketName(tmpDir, name, mesh string) string {
	return socketName(filepath.Join(tmpDir, fmt.Sprintf("kuma-mh-%s-%s", name, mesh)))
}

func socketName(s string) string {
	trimLen := len(s)
	if trimLen > 98 {
		trimLen = 98
	}
	return s[:trimLen] + ".sock"
}
