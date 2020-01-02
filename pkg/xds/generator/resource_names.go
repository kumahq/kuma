package generator

import (
	"fmt"
)

func localClusterName(port uint32) string {
	return fmt.Sprintf("localhost:%d", port)
}

func localListenerName(address string, port uint32) string {
	return fmt.Sprintf("inbound:%s:%d", address, port)
}

func envoyAdminClusterName() string {
	return "kuma:envoy:admin"
}

func prometheusListenerName() string {
	return "kuma:metrics:prometheus"
}
