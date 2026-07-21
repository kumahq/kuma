package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshMultiZoneService for generic
// resource consumers that need version-aware change detection.
func (m *MeshMultiZoneServiceResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshMultiZoneService hash used by xDS invalidation.
// Match conditions are controller bookkeeping and excluded, while the rest of
// status is hashed so newly added xDS-relevant status fields are tracked by
// default.
func (m *MeshMultiZoneServiceResource) XDSHash() []byte {
	return m.xdsHash()
}

func (m *MeshMultiZoneServiceResource) hash(includeVersion bool) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	if includeVersion {
		_, _ = hasher.Write([]byte(m.GetMeta().GetVersion()))
	}
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())

	spec := m.Spec
	if spec == nil {
		spec = &MeshMultiZoneService{}
	}
	status := m.Status
	if status == nil {
		status = &MeshMultiZoneServiceStatus{}
	}

	core_model.WriteDeterministicJSON(hasher, spec)
	core_model.WriteDeterministicJSON(hasher, status)

	return hasher.Sum(nil)
}

func (m *MeshMultiZoneServiceResource) xdsHash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())

	spec := m.Spec
	if spec == nil {
		spec = &MeshMultiZoneService{}
	}
	status := m.Status
	if status == nil {
		status = &MeshMultiZoneServiceStatus{}
	}

	core_model.WriteDeterministicJSON(hasher, spec)
	statusForHash := *status
	statusForHash.Conditions = nil
	core_model.WriteDeterministicJSON(hasher, statusForHash)
	return hasher.Sum(nil)
}
