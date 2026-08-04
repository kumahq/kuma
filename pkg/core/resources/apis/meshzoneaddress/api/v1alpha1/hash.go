package v1alpha1

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshZoneAddress for generic resource
// consumers that need version-aware change detection.
func (m *MeshZoneAddressResource) Hash() []byte {
	return m.hash(true)
}

// XDSHash returns the MeshZoneAddress hash used to gate mesh-wide xDS
// regeneration. The address is hashed because the mesh context stores it after
// DNS resolution: a load balancer hostname keeps the same resourceVersion while
// resolving to a new IP, and meta-only hashing would keep serving the stale
// endpoint. meta.GetVersion() is excluded so that writes irrelevant to xDS -
// annotations, managedFields - don't force a mesh-wide recomputation.
func (m *MeshZoneAddressResource) XDSHash() []byte {
	return m.hash(false)
}

func (m *MeshZoneAddressResource) hash(includeVersion bool) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMetaIdentity(m))
	if includeVersion {
		_, _ = hasher.Write([]byte(m.GetMeta().GetVersion()))
	}
	core_model.WriteSortedLabels(hasher, m.GetMeta().GetLabels())
	spec := m.Spec
	if spec == nil {
		spec = &MeshZoneAddress{}
	}
	core_model.WriteDeterministicJSON(hasher, spec)
	return hasher.Sum(nil)
}
