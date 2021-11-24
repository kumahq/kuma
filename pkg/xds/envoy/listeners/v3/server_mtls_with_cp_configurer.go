package v3

import (
	semver "github.com/Masterminds/semver/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideMTLSWithCPConfigurer struct {
	Ctx      xds_context.Context
	Metadata *core_xds.DataplaneMetadata
}

var _ FilterChainConfigurer = &ServerSideMTLSWithCPConfigurer{}

// backwards compatibility with 1.3.x
var HasCPValidationCtxInBootstrap = func(version *mesh_proto.Version) (bool, error) {
	if version.GetKumaDp().GetVersion() == "" { // mostly for tests but also for very old version of Kuma
		return false, nil
	}

	semverVer, err := semver.NewVersion(version.KumaDp.GetVersion())
	if err != nil {
		return false, err
	}
	return !semverVer.LessThan(semver.MustParse("1.4.0")), nil
}

func (c *ServerSideMTLSWithCPConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext, err := tls.CreateDownstreamTlsContext(c.Ctx)
	if err != nil {
		return err
	}
	if tlsContext == nil { // if mTLS is not enabled, fallback on self-signed certs
		tlsContext = tls.StaticDownstreamTlsContext(c.Ctx.ControlPlane.AdminProxyKeyPair)
	}

	hasCpValidationCtx, err := HasCPValidationCtxInBootstrap(c.Metadata.GetVersion())
	if err != nil {
		return err
	}

	if hasCpValidationCtx {
		// require certs that were presented when DP connects to the CP
		tlsContext.RequireClientCertificate = util_proto.Bool(true)
		tlsContext.CommonTlsContext.ValidationContextType = &envoy_extensions_transport_sockets_tls_v3.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: &envoy_extensions_transport_sockets_tls_v3.SdsSecretConfig{
				Name: xds_tls.CpValidationCtx,
			},
		}
	} else {
		tlsContext.RequireClientCertificate = util_proto.Bool(false)
		tlsContext.CommonTlsContext.ValidationContextType = nil
	}

	pbst, err := util_proto.MarshalAnyDeterministic(tlsContext)
	if err != nil {
		return err
	}
	filterChain.TransportSocket = &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return nil
}
