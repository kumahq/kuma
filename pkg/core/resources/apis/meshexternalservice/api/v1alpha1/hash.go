package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshExternalService for generic
// resource consumers that need version-aware change detection.
func (m *MeshExternalServiceResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshExternalService hash used by xDS invalidation.
// Hostnames and VIP assignment are hashed because they directly affect outbound
// generation and DNS domains.
func (m *MeshExternalServiceResource) XDSHash() []byte {
	return m.hash(false)
}

func (m *MeshExternalServiceResource) hash(includeVersion bool) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	if includeVersion {
		_, _ = hasher.Write([]byte(m.GetMeta().GetVersion()))
	}
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())

	spec := m.Spec
	if spec == nil {
		spec = &MeshExternalService{}
	}
	status := m.Status
	if status == nil {
		status = &MeshExternalServiceStatus{}
	}

	core_model.WriteDeterministicJSON(hasher, spec)
	core_model.WriteDeterministicJSON(hasher, status)

	return hasher.Sum(nil)
}
