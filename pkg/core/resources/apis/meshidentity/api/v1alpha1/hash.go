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

// XDSHash returns the MeshIdentity hash used by xDS invalidation. MeshIdentity
// status conditions are controller bookkeeping, except for initialization: a
// zone egress is advertised to the whole mesh only once its MeshIdentity is
// initialized, and workload identity delivery is gated on it too.
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

	if m.Status.IsInitialized() {
		_, _ = hasher.Write([]byte{1})
	} else {
		_, _ = hasher.Write([]byte{0})
	}

	if includeVersion {
		status := m.Status
		if status == nil {
			status = &MeshIdentityStatus{}
		}
		core_model.WriteDeterministicJSON(hasher, status)
	}

	return hasher.Sum(nil)
}
