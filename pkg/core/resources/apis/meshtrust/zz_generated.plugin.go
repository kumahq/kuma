package meshtrust

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	generator_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/generator/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/k8s/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.MeshTrustResourceTypeDescriptor)
	plugins.Register(plugins.PluginName(api_v1alpha1.MeshTrustResourceTypeDescriptor.KumactlArg), generator_v1alpha1.NewPlugin())
}
