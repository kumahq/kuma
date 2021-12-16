package k8s

import (
	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client_scheme "k8s.io/client-go/kubernetes/scheme"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8scnicncfio "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io"
)

// NewScheme creates a new scheme with all the necessary schemas added already (kuma CRD, builtin resources, cni CRDs...).
func NewScheme() (*kube_runtime.Scheme, error) {
	s := kube_runtime.NewScheme()
	if err := kube_client_scheme.AddToScheme(s); err != nil {
		return nil, errors.Wrapf(err, "could not add client resources to scheme")
	}
	if err := mesh_k8s.AddToScheme(s); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}
	if err := k8scnicncfio.AddToScheme(s); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", k8scnicncfio.GroupVersion)
	}
	if err := apiextensionsv1.AddToScheme(s); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", apiextensionsv1.SchemeGroupVersion)
	}
	if err := gatewayapi.Install(s); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", gatewayapi.SchemeGroupVersion)
	}
	return s, nil
}
