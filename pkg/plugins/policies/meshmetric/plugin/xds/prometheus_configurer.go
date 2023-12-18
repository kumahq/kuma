package xds

import (
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var log = core.Log.WithName("MeshMetric")

type PrometheusConfigurer struct {
	Backend         *api.PrometheusBackend
	ClusterName     string
	ListenerName    string
	EndpointAddress string
	StatsPath       string
}

func (pc *PrometheusConfigurer) ConfigureCluster(proxy *core_xds.Proxy) (envoy_common.NamedResource, error) {
	return envoy_clusters.NewClusterBuilder(proxy.APIVersion, pc.ClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(proxy.Dataplane.IsIPv6(),
			core_xds.Endpoint{
				UnixDomainPath: proxy.Metadata.MetricsSocketPath,
			},
		)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
}

func (pc *PrometheusConfigurer) ConfigureListener(proxy *core_xds.Proxy, mesh *core_mesh.MeshResource) (envoy_common.NamedResource, error) {
	var listener envoy_common.NamedResource
	var err error

	// nolint:gocritic
	if pc.useMeshMtls(mesh) {
		listener, err = pc.mtlsListener(proxy, mesh)
		if err != nil {
			return nil, err
		}
	} else if pc.useProvidedTls(proxy.Metadata) {
		listener, err = pc.providedTlsListener(proxy)
		if err != nil {
			return nil, err
		}
	} else {
		listener, err = pc.unsecuredListener(proxy)
		if err != nil {
			return nil, err
		}
	}

	return listener, nil
}

func (pc *PrometheusConfigurer) useMeshMtls(mesh *core_mesh.MeshResource) bool {
	return pc.Backend.Tls != nil && pc.Backend.Tls.Mode == api.ActiveMTLSBackend && mesh.MTLSEnabled()
}

func (pc *PrometheusConfigurer) useProvidedTls(metadata *core_xds.DataplaneMetadata) bool {
	return pc.Backend.Tls != nil && pc.Backend.Tls.Mode == api.ProvidedTLS && certsConfigured(metadata)
}

func certsConfigured(metadata *core_xds.DataplaneMetadata) bool {
	if metadata.MetricsCertPath == "" || metadata.MetricsKeyPath == "" {
		log.Info("cannot configure TLS for prometheus listener because paths to the certificate and the key wasn't provided, fallback to not secured endpoint")
		return false
	}
	return true
}

func (pc *PrometheusConfigurer) mtlsListener(proxy *core_xds.Proxy, mesh *core_mesh.MeshResource) (envoy_common.NamedResource, error) {
	match := envoy_listeners.MatchSourceAddress(proxy.Dataplane.Spec.GetNetworking().Address)
	return pc.baseSecuredListenerBuilder(proxy, match).
		Configure(envoy_listeners.FilterChain(
			envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
				envoy_listeners.ServerSideMTLS(mesh, proxy.SecretsTracker),
				envoy_listeners.StaticEndpoints(pc.ListenerName, pc.staticEndpoint()),
			),
		)).
		Build()
}

func (pc *PrometheusConfigurer) providedTlsListener(proxy *core_xds.Proxy) (envoy_common.NamedResource, error) {
	match := envoy_listeners.MatchTransportProtocol("tls")
	return pc.baseSecuredListenerBuilder(proxy, match).
		Configure(envoy_listeners.FilterChain(
			envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
				envoy_listeners.ServerSideStaticTLS(core_xds.ServerSideTLSCertPaths{
					CertPath: proxy.Metadata.MetricsCertPath,
					KeyPath:  proxy.Metadata.MetricsKeyPath,
				}),
				envoy_listeners.StaticEndpoints(pc.ListenerName, pc.staticEndpoint()),
			),
		)).
		Build()
}

func (pc *PrometheusConfigurer) unsecuredListener(proxy *core_xds.Proxy) (envoy_common.NamedResource, error) {
	return envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, pc.EndpointAddress, pc.Backend.Port, core_xds.SocketAddressProtocolTCP).
		WithOverwriteName(pc.ListenerName).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.StaticEndpoints(pc.ListenerName, pc.staticEndpoint())),
		)).
		Build()
}

func (pc *PrometheusConfigurer) baseSecuredListenerBuilder(proxy *core_xds.Proxy, match envoy_listeners.FilterChainBuilderOpt) *envoy_listeners.ListenerBuilder {
	return envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, pc.EndpointAddress, pc.Backend.Port, core_xds.SocketAddressProtocolTCP).
		WithOverwriteName(pc.ListenerName).
		// generate filter chain that does not require mTLS when DP scrapes itself (for example DP next to Prometheus Server)
		Configure(envoy_listeners.FilterChain(
			envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
				match,
				envoy_listeners.StaticEndpoints(pc.ListenerName, pc.staticEndpoint())),
		))
}

func (pc *PrometheusConfigurer) staticEndpoint() []*envoy_common.StaticEndpointPath {
	return []*envoy_common.StaticEndpointPath{
		{
			ClusterName: pc.ClusterName,
			Path:        pc.Backend.Path,
			RewritePath: pc.StatsPath,
		},
	}
}
