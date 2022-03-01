package envoy

import (
	"fmt"
	"os"

	"github.com/kumahq/kuma/pkg/core"
)

// AccessLogSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func AccessLogSocketName(name, mesh string) string {
	return socketName(fmt.Sprintf("%s%skuma-al-%s-%s", core.TempDir(), string(os.PathSeparator), name, mesh))
}

// MetricsHijackerSocketName generates a socket path that will fit the Unix socket path limitation of 104 chars
func MetricsHijackerSocketName(name, mesh string) string {
	return socketName(fmt.Sprintf("%s%skuma-mh-%s-%s", core.TempDir(), string(os.PathSeparator), name, mesh))
}

func socketName(s string) string {
	trimLen := len(s)
	if trimLen > 98 {
		trimLen = 98
	}
	return s[:trimLen] + ".sock"
}
