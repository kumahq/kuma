package meshhealthcheck

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/v1alpha1"
)

func init() {
	core.Register(
		api_v1alpha1.MeshHealthCheckResourceTypeDescriptor,
		k8s_v1alpha1.AddToScheme,
		plugin_v1alpha1.NewPlugin(),
	)
}
