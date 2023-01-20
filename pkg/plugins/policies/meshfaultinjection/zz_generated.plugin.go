package meshfaultinjection

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
	api_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/k8s/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/plugin/v1alpha1"
)

func init() {
	core.Register(
		api_v1alpha1.MeshFaultInjectionResourceTypeDescriptor,
		k8s_v1alpha1.AddToScheme,
		plugin_v1alpha1.NewPlugin(),
	)
}
