package meshproxypatch

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/plugin/v1alpha1"
)

func InitPlugin() {
	core.Register(
		api_v1alpha1.MeshProxyPatchResourceTypeDescriptor,
		k8s_v1alpha1.AddToScheme,
		plugin_v1alpha1.NewPlugin(),
	)
}
