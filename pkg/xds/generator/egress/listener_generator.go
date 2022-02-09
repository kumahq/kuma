package egress

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

// ListenerGenerator generates Zone Egress listeners.
type ListenerGenerator struct{}

var _ ZoneEgressGenerator = &ListenerGenerator{}

func (*ListenerGenerator) Generate(
	_ xds_context.Context,
	info *ResourceInfo,
) (*core_xds.ResourceSet, error) {
	if info.Resources.Listener != nil {
		return nil, nil
	}

	apiVersion := info.Proxy.APIVersion
	port := info.Listener.Port
	address := info.Listener.Address
	name := info.Listener.ResourceName

	info.Resources.Listener = envoy_listeners.NewListenerBuilder(apiVersion).
		Configure(
			envoy_listeners.InboundListener(
				name,
				address, port,
				core_xds.SocketAddressProtocolTCP,
			),
			// Always sniff for TLS.
			envoy_listeners.TLSInspector(),
		)

	return nil, nil
}
