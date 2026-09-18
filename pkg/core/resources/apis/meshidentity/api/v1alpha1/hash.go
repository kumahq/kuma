package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshIdentity for generic resource
// consumers that need version-aware change detection.
func (m *MeshIdentityResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshIdentity hash used by xDS invalidation. Status
// conditions are controller bookkeeping and are excluded, except for readiness:
// workload identity is only issued once the MeshIdentity is initialized, so the
// mesh context must be rebuilt when that flips. Without it the mesh context
// stays pinned to a not-yet-ready MeshIdentity and dataplanes never get an
// identity.
func (m *MeshIdentityResource) XDSHash() []byte {
	return m.hash(false)
}

func (m *MeshIdentityResource) hash(includeVersion bool) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	if includeVersion {
		_, _ = hasher.Write([]byte(m.GetMeta().GetVersion()))
	}
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())

	spec := m.Spec
	if spec == nil {
		spec = &MeshIdentity{}
	}
	core_model.WriteDeterministicJSON(hasher, spec)

	status := m.Status
	if status == nil {
		status = &MeshIdentityStatus{}
	}
	if includeVersion {
		core_model.WriteDeterministicJSON(hasher, status)
		return hasher.Sum(nil)
	}

	readiness := byte(0)
	if status.IsInitialized() {
		readiness = 1
	}
	_, _ = hasher.Write([]byte{readiness})

	return hasher.Sum(nil)
}
