package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideMTLSConfigurer struct {
	Mesh           *core_mesh.MeshResource
	SecretsTracker core_xds.SecretsTracker
	TlsVersion     *common_tls.Version
	TlsCiphers     common_tls.TlsCiphers
}

var _ FilterChainConfigurer = &ServerSideMTLSConfigurer{}

func (c *ServerSideMTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if !c.Mesh.MTLSEnabled() {
		return nil
	}
	tlsContext, err := tls.CreateDownstreamTlsContext(c.SecretsTracker.RequestCa(c.Mesh.GetMeta().GetName()), c.SecretsTracker.RequestIdentityCert())
	if err != nil {
		return err
	}

	version := c.TlsVersion
	if version != nil {
		if version.Min != nil {
			tlsContext.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{
				TlsMinimumProtocolVersion: common_tls.ToTlsVersion(version.Min),
			}
		}
		if version.Max != nil {
			if tlsContext.CommonTlsContext.TlsParams == nil {
				tlsContext.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{
					TlsMaximumProtocolVersion: common_tls.ToTlsVersion(version.Max),
				}
			} else {
				tlsContext.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion = common_tls.ToTlsVersion(version.Max)
			}
		}
	}

	ciphers := c.TlsCiphers
	if len(ciphers) > 0 {
		if tlsContext.CommonTlsContext.TlsParams != nil {
			var cipherSuites []string
			for _, cipher := range ciphers {
				cipherSuites = append(cipherSuites, string(cipher))
			}
			tlsContext.CommonTlsContext.TlsParams.CipherSuites = cipherSuites
		}
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
