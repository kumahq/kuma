package v1alpha1

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"strings"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	unified_naming "github.com/kumahq/kuma/v3/pkg/core/naming/unified-naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_system_names "github.com/kumahq/kuma/v3/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/mads"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/dpapi"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/metadata"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/plugin/xds"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/dynconf"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	generator_metadata "github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var (
	_                      core_plugins.PolicyPlugin = &plugin{}
	log                                              = core.Log.WithName("MeshMetric")
	DefaultRefreshInterval                           = k8s.Duration{Duration: time.Minute}
)

const (
	PrometheusListenerName       = "_kuma:metrics:prometheus"
	DefaultBackendName           = "default-backend"
	PrometheusDataplaneStatsPath = "/meshmetric"
	WorkloadAttributeKey         = "kuma.workload"
	ProxyRoleAttributeKey        = "kuma.proxy_role"

	// ProxyRole* are stable enumerated values for the kuma.proxy_role metric label.
	// Dashboards, alerts, and downstream consumers may key off these strings, so
	// treat any rename or value change as a breaking metric-contract change.
	ProxyRoleSidecar     = "sidecar"
	ProxyRoleZoneEgress  = "zone-egress"
	ProxyRoleZoneIngress = "zone-ingress"
	ProxyRoleZoneProxy   = "zone-proxy"
	ProxyRoleGateway     = "gateway"
)

func deriveProxyRole(networking *mesh_proto.Dataplane_Networking) string {
	if networking == nil {
		return ProxyRoleSidecar
	}
	if networking.GetGateway() != nil {
		return ProxyRoleGateway
	}
	if !networking.HasZoneProxyListeners() {
		return ProxyRoleSidecar
	}
	var hasIngress, hasEgress bool
	for _, l := range networking.GetListeners() {
		switch l.Type {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			hasIngress = true
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			hasEgress = true
		}
	}
	switch {
	case hasIngress && hasEgress:
		return ProxyRoleZoneProxy
	case hasIngress:
		return ProxyRoleZoneIngress
	case hasEgress:
		return ProxyRoleZoneEgress
	}
	return ProxyRoleSidecar
}

type plugin struct{}

func (p plugin) Order() int { return api.MeshMetricResourceTypeDescriptor.Order }

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

	rule := policies.SingleItemRules.Rules[0]
	policyNames := make([]string, 0, len(rule.Origin))
	for _, o := range rule.Origin {
		policyNames = append(policyNames, o.GetName())
	}
	conf := sanitizeConfForProxy(rule.Conf.(api.Conf), proxy, policyNames)

	if len(pointer.Deref(conf.Backends)) == 0 {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)
	removeResourcesConfiguredByMesh(rs, listeners.Prometheus, clusters.Prometheus)

	prometheusBackends := filterPrometheusBackends(conf.Backends)
	openTelemetryBackends := filterOpenTelemetryBackends(conf.Backends)

	unifiedNaming := unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource)
	err := configurePrometheus(rs, proxy, prometheusBackends, unifiedNaming)
	if err != nil {
		return err
	}
	if proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp) && proxy.OtelPipeBackends != nil {
		addOtelToAccumulator(proxy, openTelemetryBackends, ctx)
	}
	// configureOpenTelemetry creates Envoy-side OTel resources (listener + cluster).
	// In pipe mode only inline backends need this; addOtelToAccumulator already
	// skips them so we filter here to avoid duplicating Envoy resources for
	// backendRef backends that go through the unified pipe.
	envoyBackends := filterOtelBackendsForEnvoy(proxy, openTelemetryBackends)
	if err := configureOpenTelemetry(rs, proxy, envoyBackends, unifiedNaming, ctx.Mesh.Resources); err != nil {
		return err
	}
	inboundTagsDisabled := false
	if ctx.ControlPlane != nil {
		inboundTagsDisabled = ctx.ControlPlane.InboundTagsDisabled
	}

	if err := configureDynamicDPPConfig(rs, proxy, ctx.Mesh, conf, prometheusBackends, envoyBackends, inboundTagsDisabled); err != nil {
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

func configurePrometheus(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, prometheusBackends []*api.PrometheusBackend, unifiedNaming bool) error {
	if len(prometheusBackends) == 0 {
		return nil
	}

	for _, backend := range prometheusBackends {
		getNameOrDefault := core_system_names.GetNameOrDefault(unifiedNaming)
		backendName := pointer.DerefOr(backend.ClientId, DefaultBackendName)
		systemName := core_system_names.AsSystemName(core_system_names.JoinSections("meshmetric_prometheus", core_system_names.JoinSectionParts(core_system_names.CleanName(backendName))))

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
			StatPrefix:  getNameOrDefault(systemName, ""),
			StatsPath:   PrometheusDataplaneStatsPath,
			IPv6Enabled: proxy.Metadata.GetIPv6Enabled(),
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

func configureOpenTelemetry(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, openTelemetryBackends []*api.OpenTelemetryBackend, unifiedNaming bool, resources xds_context.Resources) error {
	for _, openTelemetryBackend := range openTelemetryBackends {
		err := configureOpenTelemetryBackend(rs, proxy, openTelemetryBackend, unifiedNaming, resources)
		if err != nil {
			return err
		}
	}
	return nil
}

func configureOpenTelemetryBackend(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, openTelemetryBackend *api.OpenTelemetryBackend, unifiedNaming bool, resources xds_context.Resources) error {
	if openTelemetryBackend == nil {
		return nil
	}

	resolved := policies_xds.ResolveOtelBackend(
		openTelemetryBackend.BackendRef,
		openTelemetryBackend.Endpoint,
		policies_xds.ParseOtelEndpoint,
		backendNameFrom,
		resources,
	)
	if resolved == nil {
		return nil
	}

	getNameOrDefault := core_system_names.GetNameOrDefault(unifiedNaming)
	endpoint := policies_xds.EndpointForDirectOtelExport(resolved, proxy.Metadata.GetDynamicMetadata(core_xds.FieldDynamicHostIP))
	backendName := resolved.Name
	systemName := core_system_names.AsSystemName(core_system_names.JoinSections("meshmetric_otel", core_system_names.JoinSectionParts(core_system_names.CleanName(backendName))))

	configurer := &plugin_xds.OpenTelemetryConfigurer{
		Endpoint:     endpoint,
		ListenerName: getNameOrDefault(systemName, envoy_names.GetOpenTelemetryListenerName(backendName)),
		ClusterName:  getNameOrDefault(systemName, envoy_names.GetOpenTelemetryListenerName(backendName)),
		StatPrefix:   getNameOrDefault(systemName, ""),
		SocketName:   core_xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir, backendName),
		ApiVersion:   proxy.APIVersion,
		IPv6Enabled:  proxy.Metadata.GetIPv6Enabled(),
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

func configureDynamicDPPConfig(
	rs *core_xds.ResourceSet,
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	conf api.Conf,
	prometheusBackends []*api.PrometheusBackend,
	openTelemetryBackends []*api.OpenTelemetryBackend,
	inboundTagsDisabled bool,
) error {
	dpConfig := createDynamicConfig(conf, proxy, meshCtx.Resource, prometheusBackends, openTelemetryBackends, inboundTagsDisabled)
	marshal, err := json.Marshal(dpConfig)
	if err != nil {
		return err
	}
	unifiedNamingEnabled := unified_naming.Enabled(proxy.Metadata, meshCtx.Resource)
	getNameOrDefault := core_system_names.GetNameOrDefault(unifiedNamingEnabled)
	return dynconf.AddConfigRoute(proxy, rs, unifiedNamingEnabled, getNameOrDefault("meshmetric", dpapi.PATH), dpapi.PATH, marshal)
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

func createDynamicConfig(
	conf api.Conf,
	proxy *core_xds.Proxy,
	mesh *core_mesh.MeshResource,
	prometheusBackends []*api.PrometheusBackend,
	openTelemetryBackends []*api.OpenTelemetryBackend,
	inboundTagsDisabled bool,
) dpapi.MeshMetricDpConfig {
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

	extraLabels := map[string]string{}
	extraLabels["mesh"] = proxy.Dataplane.GetMeta().GetMesh()
	if zone := proxy.Dataplane.GetMeta().GetLabels()[mesh_proto.ZoneTag]; zone != "" {
		extraLabels["zone"] = zone
	}
	isZoneProxyOnly := proxy.Dataplane.Spec.GetNetworking().IsZoneProxyOnly()
	extraLabels[ProxyRoleAttributeKey] = deriveProxyRole(proxy.Dataplane.Spec.GetNetworking())
	// Zone-proxy-only Dataplanes have no co-located workload, so kuma.workload is not meaningful.
	// kuma.proxy_role identifies the proxy's purpose instead.
	if !isZoneProxyOnly {
		if workloadName := proxy.Dataplane.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]; workloadName != "" {
			extraLabels[WorkloadAttributeKey] = workloadName
		}
	}
	if !unified_naming.Enabled(proxy.Metadata, mesh) {
		maps.Copy(extraLabels, mads.DataplaneLabels(proxy.Dataplane))
		extraLabels["dataplane"] = proxy.Dataplane.GetMeta().GetName()
		if extraLabels[WorkloadAttributeKey] == "" {
			if service := proxy.Dataplane.IdentifyingName(inboundTagsDisabled); service != mesh_proto.ServiceUnknown {
				extraLabels["service"] = service
			}
		}
	}

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

// filterOtelBackendsForEnvoy returns backends that need Envoy-side OTel resources.
// In pipe mode, backendRef backends go through the unified pipe so only inline
// backends need Envoy config. Without pipe mode, all backends need it.
func filterOtelBackendsForEnvoy(proxy *core_xds.Proxy, backends []*api.OpenTelemetryBackend) []*api.OpenTelemetryBackend {
	if !proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp) || proxy.OtelPipeBackends == nil {
		return backends
	}

	var inline []*api.OpenTelemetryBackend
	for _, b := range backends {
		if b != nil && b.BackendRef == nil {
			inline = append(inline, b)
		}
	}

	return inline
}

func addOtelToAccumulator(proxy *core_xds.Proxy, openTelemetryBackends []*api.OpenTelemetryBackend, ctx xds_context.Context) {
	for _, backend := range openTelemetryBackends {
		if backend == nil || backend.BackendRef == nil {
			continue
		}

		resolved := policies_xds.ResolveOtelBackend(
			backend.BackendRef,
			backend.Endpoint,
			policies_xds.ParseOtelEndpoint,
			backendNameFrom,
			ctx.Mesh.Resources,
		)
		if resolved == nil {
			continue
		}

		refreshInterval := ""
		if backend.RefreshInterval != nil {
			refreshInterval = backend.RefreshInterval.Duration.String()
		}

		base := policies_xds.BuildResolvedPipeBackend(proxy.Metadata.WorkDir, resolved)
		options := policies_xds.AddResolvedBackendOptions{
			RefreshInterval: refreshInterval,
		}
		plan := policies_xds.BuildSignalRuntimePlan(
			proxy.Metadata.GetOtelEnvInventory(),
			base.EnvPolicy,
			core_xds.OtelSignalMetrics,
			options,
		)
		proxy.OtelPipeBackends.AddSignal(resolved.Name, base, core_xds.OtelSignalMetrics, plan)
	}
}

func backendNameFrom(endpoint string) string {
	// we need to remove "/" as this name will be used as directory name
	return strings.ReplaceAll(strings.ReplaceAll(endpoint, "/", ""), ":", "-")
}

// sanitizeConfForProxy drops config fields that are meaningless on the target proxy shape.
// Applications is irrelevant on a zone-proxy-only DPP because there is no co-located
// workload to scrape — zone ingress/egress only carry mesh traffic.
// policyNames is used for log attribution only.
func sanitizeConfForProxy(conf api.Conf, proxy *core_xds.Proxy, policyNames []string) api.Conf {
	if proxy.Dataplane == nil {
		return conf
	}
	if !proxy.Dataplane.Spec.GetNetworking().IsZoneProxyOnly() {
		return conf
	}
	if len(pointer.Deref(conf.Applications)) == 0 {
		return conf
	}
	// V(1) because Apply runs on every xDS recompute; logging at default level would flood the log.
	log.V(1).Info("ignoring 'applications' on zone-proxy-only Dataplane; field has no effect without a co-located workload",
		"dataplane", proxy.Dataplane.GetMeta().GetName(),
		"mesh", proxy.Dataplane.GetMeta().GetMesh(),
		"policy", strings.Join(policyNames, ","),
	)
	conf.Applications = nil
	return conf
}
