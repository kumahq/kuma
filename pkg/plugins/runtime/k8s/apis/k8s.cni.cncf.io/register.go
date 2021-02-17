package k8scnicncfio

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	v1 "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io/v1"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "k8s.cni.cncf.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	SchemeBuilder.Register(&v1.NetworkAttachmentDefinition{}, &v1.NetworkAttachmentDefinitionList{})
}
