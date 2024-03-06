package meshmetric

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.MeshMetricResourceTypeDescriptor)
	plugins.Register(plugins.PluginName(api_v1alpha1.MeshMetricResourceTypeDescriptor.KumactlArg), plugin_v1alpha1.NewPlugin())
}
