package meshloadbalancing

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancing/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancing/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancing/plugin/v1alpha1"
)

func init() {
	core.Register(
		api_v1alpha1.MeshLoadBalancingResourceTypeDescriptor,
		k8s_v1alpha1.AddToScheme,
		plugin_v1alpha1.NewPlugin(),
	)
}
