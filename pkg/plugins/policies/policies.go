package policies

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	meshaccesslog_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshcircuitbreaker_api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshhealthcheck_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	meshratelimit_api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtrace_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var Policies = []plugins.PluginName{
	plugins.PluginName(meshaccesslog_api.MeshAccessLogResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtrace_api.MeshTraceResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshratelimit_api.MeshRateLimitResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtimeout_api.MeshTimeoutResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshtrafficpermission_api.MeshTrafficPermissionResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshcircuitbreaker_api.MeshCircuitBreakerResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshhealthcheck_api.MeshHealthCheckResourceTypeDescriptor.KumactlArg),
	plugins.PluginName(meshretry_api.MeshRetryResourceTypeDescriptor.KumactlArg),
}
