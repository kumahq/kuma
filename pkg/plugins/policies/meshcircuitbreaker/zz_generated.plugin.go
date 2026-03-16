package meshcircuitbreaker

import (
	"github.com/kumahq/kuma/v2/pkg/core/plugins"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	api_v1alpha1 "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshcircuitbreaker/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshcircuitbreaker/plugin/v1alpha1"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.MeshCircuitBreakerResourceTypeDescriptor)
	plugins.Register(plugins.PluginName(api_v1alpha1.MeshCircuitBreakerResourceTypeDescriptor.KumactlArg), plugin_v1alpha1.NewPlugin())
}
