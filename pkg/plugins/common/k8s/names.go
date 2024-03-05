package k8s

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	// K8sMeshDefaultsGenerated identifies that default resources for mesh were successfully generated
	K8sMeshDefaultsGenerated = "k8s.kuma.io/mesh-defaults-generated"

	// Kubernetes secret type to differentiate Kuma System secrets. Secret is bound to a mesh
	MeshSecretType = "system.kuma.io/secret" // #nosec G101 -- This is the name not the value

	// Kubernetes secret type to differentiate Kuma System secrets. Secret is bound to a control plane
	GlobalSecretType = "system.kuma.io/global-secret" // #nosec G101 -- This is the name not the value
)

func ResourceNameExtensions(namespace, name string) core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensions{
		core_model.K8sNamespaceComponent: namespace,
		core_model.K8sNameComponent:      name,
	}
}
