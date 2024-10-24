package names

import (
	"fmt"
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

func GetInboundRouteName(service string) string {
	return Join("inbound", service)
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

func GetDPPReadinessClusterName() string {
	return Join("_kuma", "readiness")
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

func GetOpenTelemetryListenerName(backendName string) string {
	return Join("_kuma", "metrics", "opentelemetry", backendName)
}

func GetOpenTelemetryClusterPrefix() string {
	return Join("_kuma", "metrics", "opentelemetry")
}

func GetOpenTelemetryClusterName(backendName string) string {
	return Join(GetOpenTelemetryClusterPrefix(), backendName)
}

func GetPrometheusListenerName() string {
	return Join("kuma", "metrics", "prometheus")
}

func GetAdminListenerName() string {
	return Join("kuma", "envoy", "admin")
}

func GetTracingClusterPrefix() string {
	return Join("tracing")
}

func GetTracingClusterName(backendName string) string {
	return Join(GetTracingClusterPrefix(), backendName)
}

func GetDNSListenerName() string {
	return Join("kuma", "dns")
}

func GetGatewayListenerName(gatewayName string, protoName string, port uint32) string {
	return Join(gatewayName, protoName, formatPort(port))
}

// GetMeshClusterName will be used everywhere where there is a potential of name
// clashes (i.e. when Zone Egress is configuring clusters for services with
// the same name but in different meshes)
func GetMeshClusterName(meshName string, serviceName string) string {
	return Join(meshName, serviceName)
}

// GetSecretName constructs a secret name that has a good chance of being
// unique across subsystems that are unaware of each other.
//
// category should be used to indicate the type of the secret resource. For
// example, is this a TLS certificate, or a ValidationContext, or something else.
//
// scope is a qualifier within which identifier can be considered to be unique.
// For example, the name of a Kuma file DataSource is unique across file
// DataSources, but may collide with the name of a secret DataSource.
//
// identifier is a name that should be unique within a category and scope.
func GetSecretName(category string, scope string, identifier string) string {
	return Join(category, scope, identifier)
}

func GetEgressFilterChainName(serviceName string, meshName string) string {
	return fmt.Sprintf("%s_%s", serviceName, meshName)
}
