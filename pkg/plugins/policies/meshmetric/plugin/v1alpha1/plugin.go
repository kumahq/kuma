package v1alpha1

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system_names "github.com/kumahq/kuma/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/mads"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/dpapi"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/metadata"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/dynconf"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	generator_metadata "github.com/kumahq/kuma/pkg/xds/generator/metadata"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("MeshMetric")
)

const (
	PrometheusListenerName       = "_kuma:metrics:prometheus"
	DefaultBackendName           = "default-backend"
	PrometheusDataplaneStatsPath = "/meshmetric"
	OpenTelemetryGrpcPort        = 4317
)

var DefaultRefreshInterval = k8s.Duration{Duration: time.Minute}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshMetricType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshMetricType]
	if !ok || len(policies.SingleItemRules.Rules) == 0 {
		return nil
	}

	conf := policies.SingleItemRules.Rules[0].Conf.(api.Conf)
	var kriWithoutSection kri.Identifier
	// we only handle a case where there is one origin because
	// we do not yet have a mechanism to name resources that have more than one origin https://github.com/kumahq/kuma/issues/13886
	if len(policies.SingleItemRules.Rules[0].Origin) == 1 {
		kriWithoutSection = kri.FromResourceMeta(policies.SingleItemRules.Rules[0].Origin[0], api.MeshMetricType)
	}

	if len(pointer.Deref(conf.Backends)) == 0 {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)
	removeResourcesConfiguredByMesh(rs, listeners.Prometheus, clusters.Prometheus)

	prometheusBackends := filterPrometheusBackends(conf.Backends)
	openTelemetryBackends := filterOpenTelemetryBackends(conf.Backends)

	err := configurePrometheus(rs, proxy, prometheusBackends, kriWithoutSection)
	if err != nil {
		return err
	}
	err = configureOpenTelemetry(rs, proxy, openTelemetryBackends, kriWithoutSection)
	if err != nil {
		return err
	}
	err = configureDynamicDPPConfig(rs, proxy, ctx.Mesh.Resources.MeshLocalResources, conf, prometheusBackends, openTelemetryBackends)
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

func configurePrometheus(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, prometheusBackends []*api.PrometheusBackend, kriWithoutSection kri.Identifier) error {
	if len(prometheusBackends) == 0 {
		return nil
	}

	for _, backend := range prometheusBackends {
		getNameOrDefault := core_system_names.GetNameOrDefault(proxy.Metadata.HasFeature(types.FeatureUnifiedResourceNaming) && !kriWithoutSection.IsEmpty())
		backendName := pointer.DerefOr(backend.ClientId, DefaultBackendName)
		systemName := core_system_names.AsSystemName(kri.WithSectionName(kriWithoutSection, backendName))

		configurer := &plugin_xds.PrometheusConfigurer{
			Backend: backend,
			ListenerName: getNameOrDefault(
				systemName,
				fmt.Sprintf("%s:%s", PrometheusListenerName, backendName),
			),
			EndpointAddress: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			ClusterName: getNameOrDefault(
				systemName,
				fmt.Sprintf("_%s", envoy_names.GetMetricsHijackerClusterName()),
			),
			StatsPath: PrometheusDataplaneStatsPath,
		}

		cluster, err := configurer.ConfigureCluster(proxy)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     cluster.GetName(),
			Origin:   generator_metadata.OriginPrometheus,
			Resource: cluster,
		})

		listener, err := configurer.ConfigureListener(proxy)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   generator_metadata.OriginPrometheus,
			Resource: listener,
		})
	}

	return nil
}

func configureOpenTelemetry(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, openTelemetryBackends []*api.OpenTelemetryBackend, kriWithoutSection kri.Identifier) error {
	for _, openTelemetryBackend := range openTelemetryBackends {
		err := configureOpenTelemetryBackend(rs, proxy, openTelemetryBackend, kriWithoutSection)
		if err != nil {
			return err
		}
	}
	return nil
}

func configureOpenTelemetryBackend(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, openTelemetryBackend *api.OpenTelemetryBackend, kriWithoutSection kri.Identifier) error {
	if openTelemetryBackend == nil {
		return nil
	}
	getNameOrDefault := core_system_names.GetNameOrDefault(proxy.Metadata.HasFeature(types.FeatureUnifiedResourceNaming) && !kriWithoutSection.IsEmpty())
	systemName := core_system_names.AsSystemName(kri.WithSectionName(kriWithoutSection, core_system_names.CleanName(openTelemetryBackend.Endpoint)))
	endpoint := endpointForOpenTelemetry(openTelemetryBackend.Endpoint)
	backendName := backendNameFrom(openTelemetryBackend.Endpoint)

	configurer := &plugin_xds.OpenTelemetryConfigurer{
		Endpoint:     endpoint,
		ListenerName: getNameOrDefault(systemName, envoy_names.GetOpenTelemetryListenerName(backendName)),
		ClusterName:  getNameOrDefault(systemName, envoy_names.GetOpenTelemetryListenerName(backendName)),
		SocketName:   core_xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir, backendName),
		ApiVersion:   proxy.APIVersion,
	}

	cluster, err := configurer.ConfigureCluster(proxy.Dataplane.IsIPv6())
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   metadata.OriginOpenTelemetry,
		Resource: cluster,
	})

	listener, err := configurer.ConfigureListener()
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   metadata.OriginOpenTelemetry,
		Resource: listener,
	})

	return nil
}

func configureDynamicDPPConfig(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, resources xds_context.ResourceMap, conf api.Conf, prometheusBackends []*api.PrometheusBackend, openTelemetryBackend []*api.OpenTelemetryBackend) error {
	dpConfig := createDynamicConfig(conf, proxy, resources, prometheusBackends, openTelemetryBackend)
	marshal, err := json.Marshal(dpConfig)
	if err != nil {
		return err
	}
	getNameOrDefault := core_system_names.GetNameOrDefault(proxy.Metadata.HasFeature(types.FeatureUnifiedResourceNaming))
	return dynconf.AddConfigRoute(proxy, rs, getNameOrDefault("meshmetric", dpapi.PATH), dpapi.PATH, marshal)
}

func EnvoyMetricsFilter(sidecar *api.Sidecar) url.Values {
	values := url.Values{}
	if sidecar == nil {
		values.Set("usedonly", "")
		return values
	}
	if !pointer.Deref(sidecar.IncludeUnused) {
		values.Set("usedonly", "")
	}
	return values
}

func createDynamicConfig(conf api.Conf, proxy *core_xds.Proxy, resources xds_context.ResourceMap, prometheusBackends []*api.PrometheusBackend, openTelemetryBackends []*api.OpenTelemetryBackend) dpapi.MeshMetricDpConfig {
	var applications []dpapi.Application
	for _, app := range pointer.Deref(conf.Applications) {
		applications = append(applications, dpapi.Application{
			Name:    app.Name,
			Address: pointer.Deref(app.Address),
			Port:    app.Port,
			Path:    app.Path,
		})
	}

	var backends []dpapi.Backend
	if len(prometheusBackends) != 0 {
		backends = append(backends, dpapi.Backend{
			Type: string(api.PrometheusBackendType),
		})
	}
	for _, backend := range openTelemetryBackends {
		backendName := backendNameFrom(backend.Endpoint)
		backends = append(backends, dpapi.Backend{
			Type: string(api.OpenTelemetryBackendType),
			Name: &backendName,
			OpenTelemetry: &dpapi.OpenTelemetryBackend{
				Endpoint:        core_xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir, backendName),
				RefreshInterval: pointer.DerefOr(backend.RefreshInterval, DefaultRefreshInterval),
			},
		})
	}

	var gateways []*core_mesh.MeshGatewayResource
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList).Items
	}
	extraLabels := mads.DataplaneLabels(proxy.Dataplane, gateways)

	return dpapi.MeshMetricDpConfig{
		Observability: dpapi.Observability{
			Metrics: dpapi.Metrics{
				Applications: applications,
				Backends:     backends,
				Sidecar:      conf.Sidecar,
				ExtraLabels:  extraLabels,
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

func filterOpenTelemetryBackends(backends *[]api.Backend) []*api.OpenTelemetryBackend {
	var openTelemetryBackends []*api.OpenTelemetryBackend
	for _, backend := range pointer.Deref(backends) {
		if backend.Type == api.OpenTelemetryBackendType && backend.OpenTelemetry != nil {
			openTelemetryBackends = append(openTelemetryBackends, backend.OpenTelemetry)
		}
	}
	return openTelemetryBackends
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

func backendNameFrom(endpoint string) string {
	// we need to remove "/" as this name will be used as directory name
	return strings.ReplaceAll(strings.ReplaceAll(endpoint, "/", ""), ":", "-")
}
