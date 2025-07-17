package meshidentity

import (
	api_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/k8s/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.MeshIdentityResourceTypeDescriptor)
}
