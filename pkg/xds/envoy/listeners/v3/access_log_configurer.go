package v3

import (
	"net"
	"strings"

	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

const (
	CMD_KUMA_SOURCE_ADDRESS              = "%KUMA_SOURCE_ADDRESS%"
	CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT = "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%"
	CMD_KUMA_SOURCE_SERVICE              = "%KUMA_SOURCE_SERVICE%"
	CMD_KUMA_DESTINATION_SERVICE         = "%KUMA_DESTINATION_SERVICE%"
	CMD_KUMA_MESH                        = "%KUMA_MESH%"
	CMD_KUMA_TRAFFIC_DIRECTION           = "%KUMA_TRAFFIC_DIRECTION%"
	CMD_KUMA_ZONE                        = "%KUMA_ZONE%"
	CMD_KUMA_WORKLOAD                    = "%KUMA_WORKLOAD%"
)

type KumaValues struct {
	SourceService      string
	SourceIP           string
	DestinationService string
	Mesh               string
	Zone               string
	WorkloadKRI        string
	TrafficDirection   envoy.TrafficDirection
}

const (
	cmdDownstreamRemoteAddress            = "%DOWNSTREAM_REMOTE_ADDRESS%"
	cmdDownstreamRemoteAddressWithoutPort = "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%"
)

func InterpolateKumaValues(format string, values KumaValues) string {
	sourceAddress := net.JoinHostPort(values.SourceIP, "0")
	sourceAddressWithoutPort := values.SourceIP
	if values.TrafficDirection == envoy.TrafficDirectionInbound {
		// On inbound the source is the caller, which is only known per connection,
		// so leave it to Envoy instead of rendering the proxy's own address.
		sourceAddress = cmdDownstreamRemoteAddress
		sourceAddressWithoutPort = cmdDownstreamRemoteAddressWithoutPort
	}
	format = strings.ReplaceAll(format, CMD_KUMA_SOURCE_ADDRESS, sourceAddress)
	format = strings.ReplaceAll(format, CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT, sourceAddressWithoutPort)
	format = strings.ReplaceAll(format, CMD_KUMA_SOURCE_SERVICE, values.SourceService)
	format = strings.ReplaceAll(format, CMD_KUMA_DESTINATION_SERVICE, values.DestinationService)
	format = strings.ReplaceAll(format, CMD_KUMA_MESH, values.Mesh)
	format = strings.ReplaceAll(format, CMD_KUMA_TRAFFIC_DIRECTION, string(values.TrafficDirection))
	format = strings.ReplaceAll(format, CMD_KUMA_ZONE, values.Zone)
	format = strings.ReplaceAll(format, CMD_KUMA_WORKLOAD, values.WorkloadKRI)
	return format
}
