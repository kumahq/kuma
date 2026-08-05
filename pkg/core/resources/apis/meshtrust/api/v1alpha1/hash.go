package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshTrust for generic resource
// consumers that need version-aware change detection.
func (m *MeshTrustResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshTrust hash used by xDS invalidation. Status origin
// bookkeeping does not affect trust bundles delivered to proxies.
func (m *MeshTrustResource) XDSHash() []byte {
	return m.hash(false)
}

func (m *MeshTrustResource) hash(includeVersion bool) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	if includeVersion {
		_, _ = hasher.Write([]byte(m.GetMeta().GetVersion()))
	}
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())

	spec := m.Spec
	if spec == nil {
		spec = &MeshTrust{}
	}
	core_model.WriteDeterministicJSON(hasher, spec)

	if includeVersion {
		status := m.Status
		if status == nil {
			status = &MeshTrustStatus{}
		}
		core_model.WriteDeterministicJSON(hasher, status)
	}

	return hasher.Sum(nil)
}
