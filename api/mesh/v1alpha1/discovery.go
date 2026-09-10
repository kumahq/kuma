package v1alpha1

// DiscoverySubscription describes a single ADS subscription created by a Dataplane to the
// Control Plane. Ideally there is one per Dataplane lifecycle; several indicate a lost
// connection, a Dataplane restart or a Control Plane restart.
type DiscoverySubscription struct {
	// Unique id per ADS subscription.
	Id string `json:"id,omitempty"`
	// Control Plane instance that handled the subscription.
	ControlPlaneInstanceId string `json:"controlPlaneInstanceId,omitempty"`
	// Time when the Dataplane connected to the Control Plane.
	ConnectTime *Time `json:"connectTime,omitempty"`
	// Time when the Dataplane disconnected from the Control Plane.
	DisconnectTime *Time `json:"disconnectTime,omitempty"`
	// Status of the ADS subscription.
	Status *DiscoverySubscriptionStatus `json:"status,omitempty"`
	// Version of Envoy and the Kuma dataplane.
	Version *Version `json:"version,omitempty"`
	// Generation is periodically increased by the status sink.
	Generation uint32 `json:"generation,omitempty"`
}

func (s *DiscoverySubscription) GetId() string {
	if s == nil {
		return ""
	}
	return s.Id
}

func (s *DiscoverySubscription) GetControlPlaneInstanceId() string {
	if s == nil {
		return ""
	}
	return s.ControlPlaneInstanceId
}

func (s *DiscoverySubscription) GetConnectTime() *Time {
	if s == nil {
		return nil
	}
	return s.ConnectTime
}

func (s *DiscoverySubscription) GetDisconnectTime() *Time {
	if s == nil {
		return nil
	}
	return s.DisconnectTime
}

func (s *DiscoverySubscription) GetStatus() *DiscoverySubscriptionStatus {
	if s == nil {
		return nil
	}
	return s.Status
}

func (s *DiscoverySubscription) GetVersion() *Version {
	if s == nil {
		return nil
	}
	return s.Version
}

func (s *DiscoverySubscription) GetGeneration() uint32 {
	if s == nil {
		return 0
	}
	return s.Generation
}

// DiscoverySubscriptionStatus defines the status of an ADS subscription.
type DiscoverySubscriptionStatus struct {
	// Time when the status was most recently updated.
	LastUpdateTime *Time `json:"lastUpdateTime,omitempty"`
	// Total aggregates over the individual xDS stats.
	Total *DiscoveryServiceStats `json:"total,omitempty"`
	Cds   *DiscoveryServiceStats `json:"cds,omitempty"`
	Eds   *DiscoveryServiceStats `json:"eds,omitempty"`
	Lds   *DiscoveryServiceStats `json:"lds,omitempty"`
	Rds   *DiscoveryServiceStats `json:"rds,omitempty"`
}

func (s *DiscoverySubscriptionStatus) GetLastUpdateTime() *Time {
	if s == nil {
		return nil
	}
	return s.LastUpdateTime
}

func (s *DiscoverySubscriptionStatus) GetTotal() *DiscoveryServiceStats {
	if s == nil {
		return nil
	}
	return s.Total
}

func (s *DiscoverySubscriptionStatus) GetCds() *DiscoveryServiceStats {
	if s == nil {
		return nil
	}
	return s.Cds
}

func (s *DiscoverySubscriptionStatus) GetEds() *DiscoveryServiceStats {
	if s == nil {
		return nil
	}
	return s.Eds
}

func (s *DiscoverySubscriptionStatus) GetLds() *DiscoveryServiceStats {
	if s == nil {
		return nil
	}
	return s.Lds
}

func (s *DiscoverySubscriptionStatus) GetRds() *DiscoveryServiceStats {
	if s == nil {
		return nil
	}
	return s.Rds
}

// DiscoveryServiceStats defines all stats over a single xDS service. Protobuf wrote a
// 64 bit integer as a JSON string, which the string tag reproduces.
type DiscoveryServiceStats struct {
	// Number of xDS responses sent to the Dataplane.
	// +kubebuilder:validation:Type=string
	ResponsesSent uint64 `json:"responsesSent,omitempty,string"`
	// Number of xDS responses ACKed by the Dataplane.
	// +kubebuilder:validation:Type=string
	ResponsesAcknowledged uint64 `json:"responsesAcknowledged,omitempty,string"`
	// Number of xDS responses NACKed by the Dataplane.
	// +kubebuilder:validation:Type=string
	ResponsesRejected uint64 `json:"responsesRejected,omitempty,string"`
}

func (s *DiscoveryServiceStats) GetResponsesSent() uint64 {
	if s == nil {
		return 0
	}
	return s.ResponsesSent
}

func (s *DiscoveryServiceStats) GetResponsesAcknowledged() uint64 {
	if s == nil {
		return 0
	}
	return s.ResponsesAcknowledged
}

func (s *DiscoveryServiceStats) GetResponsesRejected() uint64 {
	if s == nil {
		return 0
	}
	return s.ResponsesRejected
}

// Version defines the version of the Kuma Dataplane and Envoy.
type Version struct {
	KumaDp *KumaDpVersion `json:"kumaDp,omitempty"`
	Envoy  *EnvoyVersion  `json:"envoy,omitempty"`
	// Versions of other dependencies.
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

func (v *Version) GetKumaDp() *KumaDpVersion {
	if v == nil {
		return nil
	}
	return v.KumaDp
}

func (v *Version) GetEnvoy() *EnvoyVersion {
	if v == nil {
		return nil
	}
	return v.Envoy
}

func (v *Version) GetDependencies() map[string]string {
	if v == nil {
		return nil
	}
	return v.Dependencies
}

// KumaDpVersion describes the details of a Kuma Dataplane version.
type KumaDpVersion struct {
	Version   string `json:"version,omitempty"`
	GitTag    string `json:"gitTag,omitempty"`
	GitCommit string `json:"gitCommit,omitempty"`
	BuildDate string `json:"buildDate,omitempty"`
	// KumaCpCompatible is true when the Kuma DP version works with the CP version.
	KumaCpCompatible bool `json:"kumaCpCompatible,omitempty"`
}

func (v *KumaDpVersion) GetVersion() string {
	if v == nil {
		return ""
	}
	return v.Version
}

func (v *KumaDpVersion) GetGitTag() string {
	if v == nil {
		return ""
	}
	return v.GitTag
}

func (v *KumaDpVersion) GetGitCommit() string {
	if v == nil {
		return ""
	}
	return v.GitCommit
}

func (v *KumaDpVersion) GetBuildDate() string {
	if v == nil {
		return ""
	}
	return v.BuildDate
}

func (v *KumaDpVersion) GetKumaCpCompatible() bool {
	if v == nil {
		return false
	}
	return v.KumaCpCompatible
}

// EnvoyVersion describes the details of an Envoy version.
type EnvoyVersion struct {
	Version string `json:"version,omitempty"`
	Build   string `json:"build,omitempty"`
	// KumaDpCompatible is true when the Envoy version works with the Kuma DP version.
	KumaDpCompatible bool `json:"kumaDpCompatible,omitempty"`
}

func (v *EnvoyVersion) GetVersion() string {
	if v == nil {
		return ""
	}
	return v.Version
}

func (v *EnvoyVersion) GetBuild() string {
	if v == nil {
		return ""
	}
	return v.Build
}

func (v *EnvoyVersion) GetKumaDpCompatible() bool {
	if v == nil {
		return false
	}
	return v.KumaDpCompatible
}
