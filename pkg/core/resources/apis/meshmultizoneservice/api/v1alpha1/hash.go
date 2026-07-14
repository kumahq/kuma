package v1alpha1

import (
	"hash/fnv"

	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshMultiZoneService for generic
// resource consumers that need version-aware change detection.
func (m *MeshMultiZoneServiceResource) Hash() []byte {
	return m.hash(true, false)
}

// XDSHash returns the MeshMultiZoneService hash used by xDS invalidation.
// Match conditions are controller bookkeeping; hostnames, VIPs, and matched
// MeshServices directly affect outbound and endpoint generation.
func (m *MeshMultiZoneServiceResource) XDSHash() []byte {
	return m.hash(false, true)
}

func (m *MeshMultiZoneServiceResource) hash(includeVersion bool, xdsOnly bool) []byte {
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
	if xdsOnly {
		core_model.WriteDeterministicJSON(hasher, struct {
			Addresses          []hostnamegenerator_api.Address
			VIPs               []meshservice_api.VIP
			MeshServices       []MatchedMeshService
			HostnameGenerators []hostnamegenerator_api.HostnameGeneratorStatus
		}{
			Addresses:          status.Addresses,
			VIPs:               status.VIPs,
			MeshServices:       status.MeshServices,
			HostnameGenerators: status.HostnameGenerators,
		})
	} else {
		core_model.WriteDeterministicJSON(hasher, status)
	}

	return hasher.Sum(nil)
}
