package meshmetric

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
)

func init() {
	core.Register(
		api_v1alpha1.MeshMetricResourceTypeDescriptor,
		k8s_v1alpha1.AddToScheme,
		plugin_v1alpha1.NewPlugin(),
	)
}
