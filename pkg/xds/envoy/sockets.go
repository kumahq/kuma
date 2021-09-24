package envoy

import (
	"fmt"
	"os"
)

// AccessLogSocketName generates a socket path that will fit the Unix socket path limitation of 108 chars
func AccessLogSocketName(name, mesh string) string {
	return socketName(fmt.Sprintf("%s%skuma-al-%s-%s", os.TempDir(), string(os.PathSeparator), name, mesh))
}

// MetricsHijackerSocketName generates a socket path that will fit the Unix socket path limitation of 108 chars
func MetricsHijackerSocketName(name, mesh string) string {
	return socketName(fmt.Sprintf("%s%skuma-mh-%s-%s", os.TempDir(), string(os.PathSeparator), name, mesh))
}

func socketName(s string) string {
	trimLen := len(s)
	if trimLen > 100 {
		trimLen = 100
	}
	return s[:trimLen] + ".sock"
}
