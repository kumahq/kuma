package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideMTLSConfigurer struct {
	Mesh *core_mesh.MeshResource
}

var _ FilterChainConfigurer = &ServerSideMTLSConfigurer{}

func (c *ServerSideMTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext, err := tls.CreateDownstreamTlsContext(c.Mesh)
	if err != nil {
		return err
	}
	if tlsContext != nil {
		pbst, err := proto.MarshalAnyDeterministic(tlsContext)
		if err != nil {
			return err
		}
		filterChain.TransportSocket = &envoy_core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &envoy_core.TransportSocket_TypedConfig{
				TypedConfig: pbst,
			},
		}
	}
	return nil
}
