package names

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func GetLocalClusterName(port uint32) string {
	return fmt.Sprintf("localhost:%d", port)
}

func GetSplitClusterName(service string, idx int) string {
	return fmt.Sprintf("%s-_%d_", service, idx)
}

func GetPortForLocalClusterName(cluster string) (uint32, error) {
	parts := strings.Split(cluster, ":")
	if len(parts) != 2 {
		return 0, errors.Errorf("failed to  parse local cluster name: %s", cluster)
	}
	port, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(port), nil
}

func GetInboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("inbound:%s",
		net.JoinHostPort(address, strconv.FormatUint(uint64(port), 10)))
}

func GetOutboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("outbound:%s",
		net.JoinHostPort(address, strconv.FormatUint(uint64(port), 10)))
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

func GetMetricsHijackerClusterName() string {
	return "kuma:metrics:hijacker"
}

func GetPrometheusListenerName() string {
	return "kuma:metrics:prometheus"
}

func GetAdminListenerName() string {
	return "kuma:envoy:admin"
}

func GetTracingClusterName(backendName string) string {
	return fmt.Sprintf("tracing:%s", backendName)
}

func GetDNSListenerName() string {
	return "kuma:dns"
}

func GetGatewayListenerName(gatewayName string, protoName string, port uint32) string {
	return strings.Join([]string{gatewayName, protoName, strconv.Itoa(int(port))}, ":")
}
