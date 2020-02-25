package k8s

import (
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	// k8sNamespaceDimension identifies the namespace component of a resource name on Kubernetes.
	// The value is considered a part of user-facing Kuma API and should not be changed lightly.
	// The value has a format of a Kubernetes label name.
	k8sNamespaceDimension = "k8s.kuma.io/namespace"

	// k8sNameDimension identifies the name component of a resource name on Kubernetes.
	// The value is considered a part of user-facing Kuma API and should not be changed lightly.
	// The value has a format of a Kubernetes label name.
	k8sNameDimension = "k8s.kuma.io/name"
)

func DimensionalResourceName(namespace, name string) core_model.DimensionalResourceName {
	return core_model.DimensionalResourceName{
		k8sNamespaceDimension: namespace,
		k8sNameDimension:      name,
	}
}
