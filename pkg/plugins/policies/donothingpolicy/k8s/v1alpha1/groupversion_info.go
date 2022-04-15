// Package v1alpha1 contains API Schema definitions for the mesh v1alpha1 API group
// +groupName=kuma.io
package v1alpha1

import mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = mesh_k8s.GroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = mesh_k8s.SchemeBuilder

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = mesh_k8s.SchemeBuilder.AddToScheme
)
