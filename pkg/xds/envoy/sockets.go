package envoy

import (
	"fmt"
)

// AccessLogSocketName generates a socket path that will fit the Unix socket path limitation of 108 chars
func AccessLogSocketName(name, mesh string) string {
	return socketName(fmt.Sprintf("/tmp/kuma-al-%s-%s", name, mesh))
}

// MetricsHijackerSocketName generates a socket path that will fit the Unix socket path limitation of 108 chars
func MetricsHijackerSocketName(name, mesh string) string {
	return socketName(fmt.Sprintf("/tmp/kuma-mh-%s-%s", name, mesh))
}

func socketName(s string) string {
	trimLen := len(s)
	if trimLen > 100 {
		trimLen = 100
	}
	return s[:trimLen] + ".sock"
}
