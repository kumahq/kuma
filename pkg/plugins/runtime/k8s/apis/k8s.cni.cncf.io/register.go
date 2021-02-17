package k8scnicncfio

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	v1 "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io/v1"
)

var (
	// CNIGroupVersion is group version used to register these objects
	CNIGroupVersion = schema.GroupVersion{Group: "k8s.cni.cncf.io", Version: "v1"}

	// CNISchemeBuilder is used to add go types to the GroupVersionKind scheme
	CNISchemeBuilder = &scheme.Builder{GroupVersion: CNIGroupVersion}

	// CNIAddToScheme adds the types in this group-version to the given scheme.
	CNIAddToScheme = CNISchemeBuilder.AddToScheme

	// CRDGroupVersion is group version used to register these objects
	CRDGroupVersion = schema.GroupVersion{Group: "apiextensions.k8s.io", Version: "v1beta1"}

	// CRDSchemeBuilder is used to add go types to the GroupVersionKind scheme
	CRDSchemeBuilder = &scheme.Builder{GroupVersion: CRDGroupVersion}

	// CRDAddToScheme adds the types in this group-version to the given scheme.
	CRDAddToScheme = CRDSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	CNISchemeBuilder.Register(&v1.NetworkAttachmentDefinition{}, &v1.NetworkAttachmentDefinitionList{})
	CRDSchemeBuilder.Register(&v1beta1.CustomResourceDefinitionList{}, &v1beta1.CustomResourceDefinitionList{})
}
