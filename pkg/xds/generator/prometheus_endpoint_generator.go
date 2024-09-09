package generator

import (
	"context"
	"net"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginPrometheus is a marker to indicate by which ProxyGenerator resources were generated.
const OriginPrometheus = "prometheus"

// PrometheusEndpointGenerator generates an inbound Envoy listener
// that forwards HTTP requests into the `/stats/prometheus`
// endpoint of the Envoy Admin API.
//
// When generating such a listener, it's important not to overshadow
// a port that is already in use by the application or other Envoy listeners.
// In the latter case we prefer not generate Prometheus endpoint at all
// rather than introduce undeterministic behavior.
type PrometheusEndpointGenerator struct{}

var prometheusLog = core.Log.WithName("xds").WithName("prometheus-endpoint-generator")

func (g PrometheusEndpointGenerator) Generate(ctx context.Context, _ *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	prometheusEndpoint, err := proxy.Dataplane.GetPrometheusConfig(xdsCtx.Mesh.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "could not get prometheus endpoint")
	}
	if prometheusEndpoint == nil {
		// Prometheus metrics must be enabled Mesh-wide for Prometheus endpoint to be generated.
		return nil, nil
	}
	if proxy.Metadata.GetAdminPort() == 0 {
		// It's not possible to export Prometheus metrics if Envoy Admin API has not been enabled on that dataplane.

		// TODO(yskopets): find a way to communicate this to users
		return nil, nil
	}

	prometheusEndpointAddress := proxy.Dataplane.Spec.GetNetworking().Address
	prometheusEndpointIP := net.ParseIP(prometheusEndpointAddress)

	if proxy.Dataplane.UsesInterface(prometheusEndpointIP, prometheusEndpoint.Port) {
		// If the Prometheus endpoint would otherwise overshadow one of interfaces of that Dataplane,
		// we prefer not to do that.

		// TODO(yskopets): find a way to communicate this to users
		return nil, nil
	}

	resources := core_xds.NewResourceSet()

	statsPath := "/" + buildEnvoyMetricsFilter(prometheusEndpoint)
	metricsHijackerClusterName := envoy_names.GetMetricsHijackerClusterName()
	cluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, metricsHijackerClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(proxy.Dataplane.IsIPv6(),
			core_xds.Endpoint{
				UnixDomainPath: core_xds.MetricsHijackerSocketName(proxy.Metadata.WorkDir, proxy.Id.ToResourceKey().Name, proxy.Id.ToResourceKey().Mesh),
			},
		)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, err
	}

	resources.Add(&core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   OriginPrometheus,
		Resource: cluster,
	})

	// Cluster is generated by AdminProxyGenerator
	prometheusListenerName := envoy_names.GetPrometheusListenerName()

	var listener envoy_common.NamedResource
	if secureMetricsWithMeshMtls(prometheusEndpoint, xdsCtx.Mesh.Resource) || secureMetrics(prometheusEndpoint, proxy.Metadata) {
		var match envoy_listeners.FilterChainBuilderOpt
		switch prometheusEndpoint.Tls.GetMode() {
		case mesh_proto.PrometheusTlsConfig_activeMTLSBackend:
			match = envoy_listeners.MatchSourceAddress(proxy.Dataplane.Spec.GetNetworking().Address)
		case mesh_proto.PrometheusTlsConfig_providedTLS:
			match = envoy_listeners.MatchTransportProtocol("tls")
		}
		listenerBuilder := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, prometheusEndpointAddress, prometheusEndpoint.Port, core_xds.SocketAddressProtocolTCP).
			WithOverwriteName(prometheusListenerName).
			// generate filter chain that does not require mTLS when DP scrapes itself (for example DP next to Prometheus Server)
			Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
					match,
					envoy_listeners.StaticEndpoints(prometheusListenerName,
						[]*envoy_common.StaticEndpointPath{
							{
								ClusterName: metricsHijackerClusterName,
								Path:        prometheusEndpoint.Path,
								RewritePath: statsPath,
							},
						})),
			))

		switch prometheusEndpoint.Tls.GetMode() {
		case mesh_proto.PrometheusTlsConfig_activeMTLSBackend:
			listenerBuilder = listenerBuilder.Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
					envoy_listeners.ServerSideMTLS(xdsCtx.Mesh.Resource, proxy.SecretsTracker, nil, nil),
					envoy_listeners.StaticEndpoints(prometheusListenerName,
						[]*envoy_common.StaticEndpointPath{
							{
								ClusterName: metricsHijackerClusterName,
								Path:        prometheusEndpoint.Path,
								RewritePath: statsPath,
							},
						}),
				),
			))
		case mesh_proto.PrometheusTlsConfig_providedTLS:
			listenerBuilder = listenerBuilder.Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
					envoy_listeners.ServerSideStaticTLS(core_xds.ServerSideTLSCertPaths{
						CertPath: proxy.Metadata.MetricsCertPath,
						KeyPath:  proxy.Metadata.MetricsKeyPath,
					}),
					envoy_listeners.StaticEndpoints(prometheusListenerName,
						[]*envoy_common.StaticEndpointPath{
							{
								ClusterName: metricsHijackerClusterName,
								Path:        prometheusEndpoint.Path,
								RewritePath: statsPath,
							},
						}),
				),
			))
		}
		listener, err = listenerBuilder.Build()
	} else {
		listener, err = envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, prometheusEndpointAddress, prometheusEndpoint.Port, core_xds.SocketAddressProtocolTCP).
			WithOverwriteName(prometheusListenerName).
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
				Configure(envoy_listeners.StaticEndpoints(prometheusListenerName, []*envoy_common.StaticEndpointPath{
					{
						ClusterName: metricsHijackerClusterName,
						Path:        prometheusEndpoint.Path,
						RewritePath: statsPath,
					},
				})),
			)).
			Build()
	}
	if err != nil {
		return nil, err
	}

	resources.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   OriginPrometheus,
		Resource: listener,
	})
	return resources, nil
}

func secureMetricsWithMeshMtls(cfg *mesh_proto.PrometheusMetricsBackendConfig, mesh *core_mesh.MeshResource) bool {
	return (cfg.Tls != nil && cfg.Tls.Mode == mesh_proto.PrometheusTlsConfig_activeMTLSBackend && mesh.MTLSEnabled()) ||
		(!cfg.SkipMTLS.GetValue() && mesh.MTLSEnabled() && cfg.Tls == nil)
}

func secureMetrics(cfg *mesh_proto.PrometheusMetricsBackendConfig, metadata *core_xds.DataplaneMetadata) bool {
	if cfg.Tls != nil && cfg.Tls.GetMode() == mesh_proto.PrometheusTlsConfig_providedTLS {
		if metadata.MetricsCertPath == "" || metadata.MetricsKeyPath == "" {
			prometheusLog.Info("cannot configure TLS for prometheus listener because paths to the certificate and the key wasn't provided, fallback to not secured endpoint")
			return false
		}
		if metadata.MetricsCertPath != "" && metadata.MetricsKeyPath != "" {
			return true
		}
	}
	return false
}

// we cannot use url.Values{} because generated url looks 'usedonly='
// which isn't supported by Envoy
func buildEnvoyMetricsFilter(config *mesh_proto.PrometheusMetricsBackendConfig) string {
	var query string
	if config.GetEnvoy() != nil {
		if config.Envoy.GetFilterRegex() != "" {
			query += "filter=" + config.Envoy.GetFilterRegex()
		}
		if query != "" {
			query += "&"
		}
		if config.Envoy.GetUsedOnly().GetValue() {
			query += "usedonly"
		}
	}
	if query != "" {
		return "?" + query
	}
	return ""
}
