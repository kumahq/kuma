package gateway

import (
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

// VirtualHostGenerator generates Kuma gateway listeners.
type VirtualHostGenerator struct{}

func (*VirtualHostGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	switch p {
	case mesh_proto.Gateway_Listener_HTTP,
		mesh_proto.Gateway_Listener_HTTPS:
		return true
	default:
		return false
	}
}

func (*VirtualHostGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	info.Resources.VirtualHost = envoy_routes.NewVirtualHostBuilder(info.Proxy.APIVersion).Configure(
		// TODO(jpeach) use separator from envoy names package.
		envoy_routes.CommonVirtualHost(strings.Join([]string{info.Listener.ResourceName, info.Host.Hostname}, ":")),
		envoy_routes.DomainNames(info.Host.Hostname),
	)

	// Ensure that we get TLS on HTTPS protocol listeners.
	if info.Listener.Protocol == mesh_proto.Gateway_Listener_HTTPS {
		info.Resources.VirtualHost.Configure(envoy_routes.RequireTLS())
	}

	// TODO(jpeach) apply additional virtual host configuration.

	return nil, nil
}
