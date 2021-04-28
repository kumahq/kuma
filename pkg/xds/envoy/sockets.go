package envoy

import (
	"fmt"
)

// AccessLogSocketName generates a socket path that will fit the Unix socket path limitation of 108 chars
func AccessLogSocketName(name, mesh string) string {
	socketName := fmt.Sprintf("/tmp/kuma-al-%s-%s", name, mesh)
	trimLen := len(socketName)
	if trimLen > 100 {
		trimLen = 100
	}
	return socketName[:trimLen] + ".sock"
}

func MetricsHijackerSocketName() string {
	return "/tmp/kuma-metrics-hijacker.sock"
}
