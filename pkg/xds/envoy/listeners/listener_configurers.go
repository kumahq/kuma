package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

func TLSInspector() ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.TLSInspectorConfigurer{})
}

func OriginalDstForwarder() ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.OriginalDstForwarderConfigurer{})
}

func InboundListener(listenerName string, address string, port uint32, protocol core_xds.SocketAddressProtocol) ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.InboundListenerConfigurer{
		Protocol:     protocol,
		ListenerName: listenerName,
		Address:      address,
		Port:         port,
	})
}

func OutboundListener(listenerName string, address string, port uint32, protocol core_xds.SocketAddressProtocol) ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.OutboundListenerConfigurer{
		Protocol:     protocol,
		ListenerName: listenerName,
		Address:      address,
		Port:         port,
	})
}

func TransparentProxying(transparentProxying *mesh_proto.Dataplane_Networking_TransparentProxying) ListenerBuilderOpt {
	virtual := transparentProxying.GetRedirectPortOutbound() != 0 && transparentProxying.GetRedirectPortInbound() != 0
	if virtual {
		return AddListenerConfigurer(&v3.TransparentProxyingConfigurer{})
	}

	return ListenerBuilderOptFunc(nil)
}

func NoBindToPort() ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.TransparentProxyingConfigurer{})
}

func FilterChain(builder *FilterChainBuilder) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerConfigureFunc(func(listener *envoy_listener.Listener) error {
			filterChain, err := builder.Build()
			if err != nil {
				return err
			}
			listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener.FilterChain))
			return nil
		}),
	)
}

func DNS(vips map[string][]string, emptyDnsPort uint32, envoyVersion *mesh_proto.EnvoyVersion) ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.DNSConfigurer{
		VIPs:         vips,
		EmptyDNSPort: emptyDnsPort,
		EnvoyVersion: envoyVersion,
	})
}

func ConnectionBufferLimit(bytes uint32) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerMustConfigureFunc(func(l *envoy_listener.Listener) {
			l.PerConnectionBufferLimitBytes = wrapperspb.UInt32(bytes)
		}))
}

func EnableReusePort(enable bool) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerMustConfigureFunc(func(l *envoy_listener.Listener) {
			// TODO(jpeach) in Envoy 1.20, this field is deprecated in favor of EnableReusePort.
			l.ReusePort = enable
		}))
}

func EnableFreebind(enable bool) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerMustConfigureFunc(func(l *envoy_listener.Listener) {
			l.Freebind = wrapperspb.Bool(enable)
		}))
}
