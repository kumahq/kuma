package k8s

import (
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	// k8sNamespaceComponent identifies the namespace component of a resource name on Kubernetes.
	// The value is considered a part of user-facing Kuma API and should not be changed lightly.
	// The value has a format of a Kubernetes label name.
	k8sNamespaceComponent = "k8s.kuma.io/namespace"

	// k8sNameComponent identifies the name component of a resource name on Kubernetes.
	// The value is considered a part of user-facing Kuma API and should not be changed lightly.
	// The value has a format of a Kubernetes label name.
	k8sNameComponent = "k8s.kuma.io/name"
)

func ResourceNameExtensions(namespace, name string) core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensions{
		k8sNamespaceComponent: namespace,
		k8sNameComponent:      name,
	}
}
