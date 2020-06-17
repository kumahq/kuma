package names

import (
	"fmt"
)

func GetLocalClusterName(port uint32) string {
	return fmt.Sprintf("localhost:%d", port)
}

func GetInboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("inbound:%s:%d", address, port)
}

func GetOutboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("outbound:%s:%d", address, port)
}

func GetInboundRouteName(service string) string {
	return fmt.Sprintf("inbound:%s", service)
}

func GetOutboundRouteName(service string) string {
	return fmt.Sprintf("outbound:%s", service)
}

func GetEnvoyAdminClusterName() string {
	return "kuma:envoy:admin"
}

func GetPrometheusListenerName() string {
	return "kuma:metrics:prometheus"
}
