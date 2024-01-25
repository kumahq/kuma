package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("MeshMetric")
)

const (
	OriginOpenTelemetry       = "open-telemetry"
	OriginDynamicConfig       = "dynamic-config"
	PrometheusListenerName    = "_kuma:metrics:prometheus"
	DynamicConfigListenerName = "_kuma:dynamicconfig:observability"
	DefaultBackendName        = "default-backend"
	OpenTelemetryGrpcPort     = 4317
)

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshMetricType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshMetricType]
	if !ok || len(policies.SingleItemRules.Rules) == 0 {
		return nil
	}

	conf := policies.SingleItemRules.Rules[0].Conf.(api.Conf)

	if len(pointer.Deref(conf.Backends)) == 0 {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)
	removeResourcesConfiguredByMesh(rs, listeners.Prometheus, clusters.Prometheus)

	prometheusBackends := filterPrometheusBackends(conf.Backends)
	// TODO multiple backends of the same type support. Issue: https://github.com/kumahq/kuma/issues/8942
	openTelemetryBackend := firstOpenTelemetryBackend(conf.Backends)

	err := configurePrometheus(rs, proxy, prometheusBackends, conf)
	if err != nil {
		return err
	}
	err = configureOpenTelemetry(rs, proxy, openTelemetryBackend)
	if err != nil {
		return err
	}
	err = configureDynamicDPPConfig(rs, proxy, conf, prometheusBackends, openTelemetryBackend)
	if err != nil {
		return err
	}

	return nil
}

func removeResourcesConfiguredByMesh(rs *core_xds.ResourceSet, listener *envoy_listener.Listener, cluster *envoy_cluster.Cluster) {
	if cluster != nil && listener != nil {
		log.Info("You should not use MeshMetric policy together with metrics configured in Mesh. MeshMetric will take precedence over Mesh configuration")
		rs.Remove(envoy_resource.ClusterType, cluster.Name)
		rs.Remove(envoy_resource.ListenerType, listener.Name)
	}
}

func configurePrometheus(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, prometheusBackends []*api.PrometheusBackend, conf api.Conf) error {
	if len(prometheusBackends) == 0 {
		return nil
	}

	for _, backend := range prometheusBackends {
		configurer := &plugin_xds.PrometheusConfigurer{
			Backend:         backend,
			ListenerName:    fmt.Sprintf("%s:%s", PrometheusListenerName, pointer.DerefOr(backend.ClientId, DefaultBackendName)),
			EndpointAddress: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			ClusterName:     fmt.Sprintf("_%s", envoy_names.GetMetricsHijackerClusterName()),
			StatsPath:       "/" + envoyMetricsFilter(conf),
		}

		cluster, err := configurer.ConfigureCluster(proxy)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     cluster.GetName(),
			Origin:   generator.OriginPrometheus,
			Resource: cluster,
		})

		listener, err := configurer.ConfigureListener(proxy)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   generator.OriginPrometheus,
			Resource: listener,
		})
	}

	return nil
}

func configureOpenTelemetry(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, openTelemetryBackend *api.OpenTelemetryBackend) error {
	if openTelemetryBackend == nil {
		return nil
	}
	endpoint := endpointForOpenTelemetry(openTelemetryBackend.Endpoint)

	configurer := &plugin_xds.OpenTelemetryConfigurer{
		Endpoint:     endpoint,
		ListenerName: "_kuma:metrics:opentelemetry",
		ClusterName:  "_kuma:metrics:opentelemetry:collector",
	}

	cluster, err := configurer.ConfigureCluster(proxy)
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   OriginOpenTelemetry,
		Resource: cluster,
	})

	listener, err := configurer.ConfigureListener(proxy)
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   OriginOpenTelemetry,
		Resource: listener,
	})

	return nil
}

func configureDynamicDPPConfig(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, conf api.Conf, prometheusBackends []*api.PrometheusBackend, openTelemetryBackend *api.OpenTelemetryBackend) error {
	configurer := &plugin_xds.DppConfigConfigurer{
		ListenerName: DynamicConfigListenerName,
		DpConfig:     createDynamicConfig(conf, proxy, prometheusBackends, openTelemetryBackend),
	}

	listener, err := configurer.ConfigureListener(proxy)
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   OriginDynamicConfig,
		Resource: listener,
	})

	return nil
}

// TODO this most likely won't work with OpenTelemetry. Issue: https://github.com/kumahq/kuma/issues/8926
func envoyMetricsFilter(conf api.Conf) string {
	if conf.Sidecar == nil {
		return "?usedonly" // as the default for IncludeUnused is false
	}
	var query string
	if pointer.Deref(conf.Sidecar.Regex) != "" {
		query += "filter=" + pointer.Deref(conf.Sidecar.Regex)
	}
	if query != "" {
		query += "&"
	}
	if !pointer.Deref(conf.Sidecar.IncludeUnused) {
		query += "usedonly"
	}
	if query != "" {
		return "?" + query
	}
	return ""
}

func createDynamicConfig(conf api.Conf, proxy *core_xds.Proxy, prometheusBackends []*api.PrometheusBackend, openTelemetryBackend *api.OpenTelemetryBackend) plugin_xds.MeshMetricDpConfig {
	var applications []plugin_xds.Application
	for _, app := range pointer.Deref(conf.Applications) {
		applications = append(applications, plugin_xds.Application{
			Name:    app.Name,
			Address: pointer.Deref(app.Address),
			Port:    app.Port,
			Path:    pointer.DerefOr(app.Path, "/metrics"),
		})
	}

	var backends []plugin_xds.Backend
	if len(prometheusBackends) != 0 {
		backends = append(backends, plugin_xds.Backend{
			Type: string(api.PrometheusBackendType),
		})
	}
	if openTelemetryBackend != nil {
		backends = append(backends, plugin_xds.Backend{
			Type: string(api.OpenTelemetryBackendType),
			OpenTelemetry: &plugin_xds.OpenTelemetryBackend{
				Endpoint: core_xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir),
			},
		})
	}

	return plugin_xds.MeshMetricDpConfig{
		Observability: plugin_xds.Observability{
			Metrics: plugin_xds.Metrics{
				Applications: applications,
				Backends:     backends,
			},
		},
	}
}

func endpointForOpenTelemetry(endpoint string) *core_xds.Endpoint {
	target := strings.Split(endpoint, ":")
	port := uint32(OpenTelemetryGrpcPort) // default gRPC port
	if len(target) > 1 {
		val, _ := strconv.ParseInt(target[1], 10, 32)
		port = uint32(val)
	}
	return &core_xds.Endpoint{
		Target: target[0],
		Port:   port,
	}
}

func firstOpenTelemetryBackend(backends *[]api.Backend) *api.OpenTelemetryBackend {
	for _, backend := range pointer.Deref(backends) {
		if backend.Type == api.OpenTelemetryBackendType && backend.OpenTelemetry != nil {
			return backend.OpenTelemetry
		}
	}
	return nil
}

func filterPrometheusBackends(backends *[]api.Backend) []*api.PrometheusBackend {
	var prometheusBackends []*api.PrometheusBackend
	for _, backend := range pointer.Deref(backends) {
		if backend.Type == api.PrometheusBackendType && backend.Prometheus != nil {
			prometheusBackends = append(prometheusBackends, backend.Prometheus)
		}
	}
	return prometheusBackends
}
