package meshzoneaddress

import (
	api_v1alpha1 "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	k8s_v1alpha1 "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
)

func InitPlugin() {
	registry.AddKubeScheme(k8s_v1alpha1.AddToScheme)
	registry.RegisterType(api_v1alpha1.MeshZoneAddressResourceTypeDescriptor)
}
