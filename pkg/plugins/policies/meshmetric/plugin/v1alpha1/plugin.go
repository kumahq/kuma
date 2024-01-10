package v1alpha1

import (
	"fmt"

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
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/xds"
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
	OriginDynamicConfig       = "dynamic-config"
	DynamicConfigListenerName = "_kuma:dynamicconfig:observability"
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

	backend := pointer.Deref(conf.Backends)[0]
	if backend.Type == api.PrometheusBackendType && backend.Prometheus != nil {
		err := configurePrometheus(rs, proxy, backend.Prometheus, conf)
		if err != nil {
			return err
		}
	}

	err := configureDynamicDPPConfig(rs, proxy, conf)
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

func configurePrometheus(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, prometheusBackend *api.PrometheusBackend, conf api.Conf) error {
	configurer := &xds.PrometheusConfigurer{
		Backend:         prometheusBackend,
		ListenerName:    fmt.Sprintf("_%s", envoy_names.GetPrometheusListenerName()),
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

	return nil
}

func configureDynamicDPPConfig(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, conf api.Conf) error {
	configurer := &xds.DppConfigConfigurer{
		ListenerName: DynamicConfigListenerName,
		DpConfig:     createDynamicConfig(conf),
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

func envoyMetricsFilter(conf api.Conf) string {
	if conf.Sidecar == nil {
		return ""
	}
	var query string
	if pointer.Deref(conf.Sidecar.Regex) != "" {
		query += "filter=" + pointer.Deref(conf.Sidecar.Regex)
	}
	if query != "" {
		query += "&"
	}
	if pointer.Deref(conf.Sidecar.UsedOnly) {
		query += "usedonly"
	}
	if query != "" {
		return "?" + query
	}
	return ""
}

func createDynamicConfig(conf api.Conf) xds.MeshMetricDpConfig {
	var applications []xds.Application
	for _, app := range pointer.Deref(conf.Applications) {
		applications = append(applications, xds.Application{
			Port: app.Port,
			Path: pointer.DerefOr(app.Path, "/metrics"),
		})
	}

	return xds.MeshMetricDpConfig{
		Observability: xds.Observability{
			Metrics: xds.Metrics{
				Applications: applications,
			},
		},
	}
}
