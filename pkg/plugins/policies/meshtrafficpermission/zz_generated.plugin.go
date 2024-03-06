package meshtrafficpermission

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/plugin/v1alpha1"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)
	plugins.Register(plugins.PluginName(api_v1alpha1.MeshTrafficPermissionResourceTypeDescriptor.KumactlArg), plugin_v1alpha1.NewPlugin())
}
