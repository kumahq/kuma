package v1alpha1

// MeshInsight defines the observed state of a Mesh.
type MeshInsight struct {
	// Dataplanes aggregates over the dataplanes of the mesh.
	Dataplanes *MeshInsight_DataplaneStat `json:"dataplanes,omitempty"`
	// DpVersions holds the dataplane stats grouped by version.
	DpVersions *MeshInsight_DpVersions `json:"dpVersions,omitempty"`
	// MTLS holds the dataplane stats grouped by certificate backend.
	MTLS *MeshInsight_MTLS `json:"mTLS,omitempty"`
	// DataplanesByType splits the dataplane stats by proxy type.
	DataplanesByType *MeshInsight_DataplanesByType `json:"dataplanesByType,omitempty"`
	// Resources counts the stored resources of the mesh by type.
	Resources map[string]*MeshInsight_ResourceStat `json:"resources,omitempty"`
}

func (m *MeshInsight) GetDataplanes() *MeshInsight_DataplaneStat {
	if m == nil {
		return nil
	}
	return m.Dataplanes
}

func (m *MeshInsight) GetDpVersions() *MeshInsight_DpVersions {
	if m == nil {
		return nil
	}
	return m.DpVersions
}

func (m *MeshInsight) GetMTLS() *MeshInsight_MTLS {
	if m == nil {
		return nil
	}
	return m.MTLS
}

func (m *MeshInsight) GetDataplanesByType() *MeshInsight_DataplanesByType {
	if m == nil {
		return nil
	}
	return m.DataplanesByType
}

func (m *MeshInsight) GetResources() map[string]*MeshInsight_ResourceStat {
	if m == nil {
		return nil
	}
	return m.Resources
}

// MeshInsight_DataplaneStat defines statistics specifically for Dataplanes.
type MeshInsight_DataplaneStat struct {
	Total uint32 `json:"total,omitempty"`
	// Online counts the dataplanes that are connected.
	Online uint32 `json:"online,omitempty"`
	// Offline counts the dataplanes that are not connected.
	Offline uint32 `json:"offline,omitempty"`
	// PartiallyDegraded counts the dataplanes with some inbounds offline.
	PartiallyDegraded uint32 `json:"partiallyDegraded,omitempty"`
}

func (s *MeshInsight_DataplaneStat) GetTotal() uint32 {
	if s == nil {
		return 0
	}
	return s.Total
}

func (s *MeshInsight_DataplaneStat) GetOnline() uint32 {
	if s == nil {
		return 0
	}
	return s.Online
}

func (s *MeshInsight_DataplaneStat) GetOffline() uint32 {
	if s == nil {
		return 0
	}
	return s.Offline
}

func (s *MeshInsight_DataplaneStat) GetPartiallyDegraded() uint32 {
	if s == nil {
		return 0
	}
	return s.PartiallyDegraded
}

// MeshInsight_DpVersions defines statistics grouped by dataplane versions.
type MeshInsight_DpVersions struct {
	// KumaDp holds the dataplane stats grouped by KumaDP version.
	KumaDp map[string]*MeshInsight_DataplaneStat `json:"kumaDp,omitempty"`
	// Envoy holds the dataplane stats grouped by Envoy version.
	Envoy map[string]*MeshInsight_DataplaneStat `json:"envoy,omitempty"`
}

func (v *MeshInsight_DpVersions) GetKumaDp() map[string]*MeshInsight_DataplaneStat {
	if v == nil {
		return nil
	}
	return v.KumaDp
}

func (v *MeshInsight_DpVersions) GetEnvoy() map[string]*MeshInsight_DataplaneStat {
	if v == nil {
		return nil
	}
	return v.Envoy
}

// MeshInsight_MTLS groups the dataplanes by certificate backend.
type MeshInsight_MTLS struct {
	// IssuedBackends groups the dataplanes by the backend that issued their
	// certificate.
	IssuedBackends map[string]*MeshInsight_DataplaneStat `json:"issuedBackends,omitempty"`
	// SupportedBackends groups the dataplanes by the backends they accept
	// certificates from.
	SupportedBackends map[string]*MeshInsight_DataplaneStat `json:"supportedBackends,omitempty"`
}

func (m *MeshInsight_MTLS) GetIssuedBackends() map[string]*MeshInsight_DataplaneStat {
	if m == nil {
		return nil
	}
	return m.IssuedBackends
}

func (m *MeshInsight_MTLS) GetSupportedBackends() map[string]*MeshInsight_DataplaneStat {
	if m == nil {
		return nil
	}
	return m.SupportedBackends
}

// MeshInsight_DataplanesByType splits the statistics by dataplane type.
type MeshInsight_DataplanesByType struct {
	// Standard holds the stats of every dataplane.
	Standard *MeshInsight_DataplaneStat `json:"standard,omitempty"`
}

func (d *MeshInsight_DataplanesByType) GetStandard() *MeshInsight_DataplaneStat {
	if d == nil {
		return nil
	}
	return d.Standard
}

// MeshInsight_ResourceStat counts the stored resources of a single type.
type MeshInsight_ResourceStat struct {
	Total uint32 `json:"total,omitempty"`
}

func (r *MeshInsight_ResourceStat) GetTotal() uint32 {
	if r == nil {
		return 0
	}
	return r.Total
}
