package k8s

import (
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8scnicncfio "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io"
)

// NewScheme creates a new scheme with all the necessary schemas added already (kuma CRD, builtin resources, cni CRDs...).
func NewScheme() (*kube_runtime.Scheme, error) {
	s := kube_runtime.NewScheme()
	if err := kube_core.AddToScheme(s); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
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
	if err := appsv1.AddToScheme(s); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", apiextensionsv1.SchemeGroupVersion)
	}
	return s, nil
}
