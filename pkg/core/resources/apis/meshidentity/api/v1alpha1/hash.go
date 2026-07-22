package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshIdentity for generic resource
// consumers that need version-aware change detection.
func (m *MeshIdentityResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshIdentity hash used by xDS invalidation. MeshIdentity
// status conditions are controller bookkeeping and do not affect config
// generation or selected workload identity contents. Per-dataplane identity
// delivery still tracks the readiness flip separately in
// pkg/xds/sync/dataplane_watchdog.go:hashMeshIdentity.
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

	if includeVersion {
		status := m.Status
		if status == nil {
			status = &MeshIdentityStatus{}
		}
		core_model.WriteDeterministicJSON(hasher, status)
	}

	return hasher.Sum(nil)
}
