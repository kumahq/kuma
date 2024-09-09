package ordered

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	meshaccesslog_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshcircuitbreaker_api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshfaultinjection_api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	meshhealthcheck_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshloadbalancingstrategy_api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	meshmetric_api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshpassthrough_api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	meshproxypatch_api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	meshratelimit_api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtls_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	meshtrace_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var Policies = []plugins.PluginName{
	// Routes have to come first
	plugins.PluginName(meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtcproute_api.MeshTCPRouteResourceTypeDescriptor.KumactlArg),
	// MeshTLS needs to come before everything because it rebuilds the inbound listeners
	plugins.PluginName(meshtls_api.MeshTLSResourceTypeDescriptor.KumactlArg),
	// For other policies order isn't important at the moment
	plugins.PluginName(meshloadbalancingstrategy_api.MeshLoadBalancingStrategyResourceTypeDescriptor.KumactlArg),
	// has to be before MeshAccessLog so the plugin can access log filters that are added to the filter chains
	plugins.PluginName(meshpassthrough_api.MeshPassthroughResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshaccesslog_api.MeshAccessLogResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtrace_api.MeshTraceResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshfaultinjection_api.MeshFaultInjectionResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshratelimit_api.MeshRateLimitResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtimeout_api.MeshTimeoutResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtrafficpermission_api.MeshTrafficPermissionResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshcircuitbreaker_api.MeshCircuitBreakerResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshhealthcheck_api.MeshHealthCheckResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshretry_api.MeshRetryResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshmetric_api.MeshMetricResourceTypeDescriptor.KumactlArg),
	// MeshProxyPatch comes after all others
	plugins.PluginName(meshproxypatch_api.MeshProxyPatchResourceTypeDescriptor.KumactlArg),
}
