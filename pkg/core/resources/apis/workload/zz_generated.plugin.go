package workload

import (
	api_v1alpha1 "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.WorkloadResourceTypeDescriptor)
}
