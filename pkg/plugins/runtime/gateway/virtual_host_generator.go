package gateway

import (
	"strings"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
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

func (*VirtualHostGenerator) Generate(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	if info.Resources.VirtualHost != nil {
		return nil, nil
	}

	info.Resources.VirtualHost = envoy_routes.NewVirtualHostBuilder(info.Proxy.APIVersion).Configure(
		// TODO(jpeach) use separator from envoy names package.
		envoy_routes.CommonVirtualHost(strings.Join([]string{info.Listener.ResourceName, info.Host.Hostname}, ":")),
		envoy_routes.DomainNames(info.Host.Hostname),
	)

	// Ensure that we get TLS on HTTPS protocol listeners.
	if info.Listener.Protocol == mesh_proto.Gateway_Listener_HTTPS {
		info.Resources.VirtualHost.Configure(
			envoy_routes.AddVirtualHostConfigurer(
				v3.VirtualHostMustConfigureFunc(func(vh *envoy_route.VirtualHost) {
					vh.RequireTls = envoy_route.VirtualHost_ALL
				}),
			),
		)
	}

	// TODO(jpeach) apply additional virtual host configuration.

	return nil, nil
}
