package names

import (
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Separator is the separator used in resource names.
const Separator = ":"

func formatPort(port uint32) string {
	return strconv.FormatUint(uint64(port), 10)
}

// Join uses Separator to join the given parts into a resource name.
func Join(parts ...string) string {
	return strings.Join(parts, Separator)
}

// Renaming might break metrics
// https://github.com/kumahq/kuma/issues/3249
func GetLocalClusterName(port uint32) string {
	return Join("localhost", formatPort(port))
}

func GetPortForLocalClusterName(cluster string) (uint32, error) {
	parts := strings.Split(cluster, Separator)
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
	return Join("inbound",
		net.JoinHostPort(address, formatPort(port)))
}

func GetOutboundListenerName(address string, port uint32) string {
	return Join("outbound",
		net.JoinHostPort(address, formatPort(port)))
}

func GetOutboundRouteName(service string) string {
	return Join("outbound", service)
}

func GetEnvoyAdminClusterName() string {
	return Join("kuma", "envoy", "admin")
}

func GetMetricsHijackerClusterName() string {
	return Join("kuma", "metrics", "hijacker")
}

func GetInternalClusterNamePrefix() string {
	return "_"
}

func GetAdsClusterName() string {
	return "ads_cluster"
}

func GetAccessLogSinkClusterName() string {
	return "access_log_sink"
}

func GetOpenTelemetryClusterPrefix() string {
	return Join("_kuma", "metrics", "opentelemetry")
}
