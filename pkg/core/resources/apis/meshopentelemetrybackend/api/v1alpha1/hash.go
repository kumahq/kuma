package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshOpenTelemetryBackend for generic
// consumers that need version-aware change detection.
func (m *MeshOpenTelemetryBackendResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshOpenTelemetryBackend hash used by xDS invalidation.
// Status conditions are observability-only and do not affect generated config.
func (m *MeshOpenTelemetryBackendResource) XDSHash() []byte {
	return m.hash(false)
}

func (m *MeshOpenTelemetryBackendResource) hash(includeVersion bool) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	if includeVersion {
		_, _ = hasher.Write([]byte(m.GetMeta().GetVersion()))
	}
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())

	spec := m.Spec
	if spec == nil {
		spec = &MeshOpenTelemetryBackend{}
	}
	core_model.WriteDeterministicJSON(hasher, spec)

	if includeVersion {
		status := m.Status
		if status == nil {
			status = &MeshOpenTelemetryBackendStatus{}
		}
		core_model.WriteDeterministicJSON(hasher, status)
	}

	return hasher.Sum(nil)
}
