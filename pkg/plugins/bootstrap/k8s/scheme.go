package k8s

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client_scheme "k8s.io/client-go/kubernetes/scheme"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kumahq/kuma/pkg/plugins/policies"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8scnicncfio "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io"
)

// NewScheme creates a new scheme with all the necessary schemas added already (kuma CRD, builtin resources, cni CRDs...).
func NewScheme() (*kube_runtime.Scheme, error) {
	s := kube_runtime.NewScheme()
	if err := kube_client_scheme.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("could not add client resources to scheme: %w", err)
	}
	if err := mesh_k8s.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("could not add %q to scheme: %w", mesh_k8s.GroupVersion, err)
	}
	if err := k8scnicncfio.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("could not add %q to scheme: %w", k8scnicncfio.GroupVersion, err)
	}
	if err := apiextensionsv1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("could not add %q to scheme: %w", apiextensionsv1.SchemeGroupVersion, err)
	}
	if err := gatewayapi.Install(s); err != nil {
		return nil, fmt.Errorf("could not add %q to scheme: %w", gatewayapi.SchemeGroupVersion, err)
	}
	if err := policies.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}
