package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
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

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshMetricType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	log.V(1).Info("apply is not implemented")
	policies, ok := proxy.Policies.Dynamic[api.MeshMetricType]
	if !ok || len(policies.SingleItemRules.Rules) == 0 {
		return nil
	}

	conf := policies.SingleItemRules.Rules[0].Conf.(api.Conf)

	if len(pointer.Deref(conf.Backends)) == 0 {
		return nil
	}

	backend := pointer.Deref(conf.Backends)[0]
	if backend.Type == api.PrometheusBackendType && backend.Prometheus != nil {
		err := configurePrometheus(rs, proxy, ctx.Mesh.Resource, backend.Prometheus, conf)
		if err != nil {
			return err
		}
	}

	return nil
}

func configurePrometheus(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, mesh *core_mesh.MeshResource, prometheusBackend *api.PrometheusBackend, conf api.Conf) error {
	configurer := &xds.PrometheusConfigurer{
		Backend:         prometheusBackend,
		ListenerName:    envoy_names.GetPrometheusListenerName(),
		EndpointAddress: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
		ClusterName:     envoy_names.GetMetricsHijackerClusterName(),
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

	listener, err := configurer.ConfigureListener(proxy, mesh)
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
