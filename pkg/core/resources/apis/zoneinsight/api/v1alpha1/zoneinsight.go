// +kubebuilder:object:generate=true
package v1alpha1

import "time"

// ZoneInsight defines the observed state of a Zone Kuma CP.
// +kuma:policy:is_policy=false
// +kuma:policy:scope=Global
// +kuma:policy:cluster_scoped_k8s=true
// +kuma:policy:read_only=true
// +kuma:policy:skip_kumactl=true
// +kuma:policy:policy_matching_exempt=true
// +kuma:policy:opaque_k8s_spec=true
// +kuma:policy:plugin_originated=false
// +kuma:policy:short_name=
// +kuma:policy:ws_name=zone-insight
// +kuma:policy:kds_flags=model.ProvidedByGlobalFlag | model.ProvidedByZoneFlag
type ZoneInsight struct {
	// List of KDS subscriptions created by a given Zone Kuma CP.
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	Subscriptions []*KDSSubscription `json:"subscriptions,omitempty"`
	// Statistics about Envoy Admin Streams, superseded by KDSStreams.
	EnvoyAdminStreams *EnvoyAdminStreams `json:"envoyAdminStreams,omitempty"`
	// Information about the last received zone health check.
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
	// Information about kds streams that are established between global and zone.
	KDSStreams *KDSStreams `json:"kdsStreams,omitempty"`
}

type EnvoyAdminStreams struct {
	// Global instance ID that handles XDS Config Dump streams.
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	ConfigDumpGlobalInstanceID string `json:"configDumpGlobalInstanceId,omitempty"`
	// Global instance ID that handles Stats streams.
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	StatsGlobalInstanceID string `json:"statsGlobalInstanceId,omitempty"`
	// Global instance ID that handles Clusters streams.
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	ClustersGlobalInstanceID string `json:"clustersGlobalInstanceId,omitempty"`
}

type KDSStreams struct {
	// Details of stream that handles Clusters stream.
	Clusters *KDSStream `json:"clusters,omitempty"`
	// Details of stream that handles XDS Config Dump stream.
	ConfigDump *KDSStream `json:"configDump,omitempty"`
	// Details of stream that handles Stats stream.
	Stats *KDSStream `json:"stats,omitempty"`
	// Details of stream that handles global to zone resource sync stream.
	GlobalToZone *KDSStream `json:"globalToZone,omitempty"`
	// Details of stream that handles zone to global resource sync stream.
	ZoneToGlobal *KDSStream `json:"zoneToGlobal,omitempty"`
}

type KDSStream struct {
	// Global instance ID that handles the stream.
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	GlobalInstanceID string `json:"globalInstanceId,omitempty"`
	// Time when the stream was open.
	ConnectTime *Time `json:"connectTime,omitempty"`
}

// KDSSubscription describes a single KDS subscription created by a Zone to the Global.
// Ideally there is one per Zone lifecycle. Several of them indicate a transient loss of
// connectivity, or a restart of either control plane.
type KDSSubscription struct {
	// Unique id per KDS subscription.
	ID string `json:"id,omitempty"`
	// Global CP instance that handled given subscription.
	GlobalInstanceID string `json:"globalInstanceId,omitempty"`
	// Time when a given Zone connected to the Global.
	ConnectTime *Time `json:"connectTime,omitempty"`
	// Time when a given Zone disconnected from the Global.
	DisconnectTime *Time `json:"disconnectTime,omitempty"`
	// Status of the KDS subscription.
	Status *KDSSubscriptionStatus `json:"status,omitempty"`
	// Version of Zone Kuma CP.
	Version *Version `json:"version,omitempty"`
	// Generation is an integer number which is periodically increased by the status sink.
	Generation uint32 `json:"generation,omitempty"`
	// Config of Zone Kuma CP.
	Config string `json:"config,omitempty"`
	// Indicates if subscription provided auth token.
	AuthTokenProvided bool `json:"authTokenProvided,omitempty"`
	// Zone CP instance that handled the given subscription, the leader at the time of
	// connection.
	ZoneInstanceID string `json:"zoneInstanceId,omitempty"`
}

type KDSSubscriptionStatus struct {
	// Time when status of a given KDS subscription was most recently updated.
	LastUpdateTime *Time `json:"lastUpdateTime,omitempty"`
	// Total defines an aggregate over individual KDS stats.
	Total *KDSServiceStats `json:"total,omitempty"`
	// Stat holds the KDS stats per resource type.
	Stat map[string]*KDSServiceStats `json:"stat,omitempty"`
}

// KDSServiceStats defines all stats over a single xDS service. The counters serialize
// as JSON strings because protobuf writes 64 bit integers that way, and an older
// control plane rejects them when written as numbers.
type KDSServiceStats struct {
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

// Version defines version of Kuma ControlPlane.
type Version struct {
	// Version of Zone Kuma CP.
	KumaCP *KumaCpVersion `json:"kumaCp,omitempty"`
}

// KumaCpVersion describes details of Kuma ControlPlane version.
type KumaCpVersion struct {
	// Version number of Kuma ControlPlane.
	Version string `json:"version,omitempty"`
	// Git tag of Kuma ControlPlane version.
	GitTag string `json:"gitTag,omitempty"`
	// Git commit of Kuma ControlPlane version.
	GitCommit string `json:"gitCommit,omitempty"`
	// Build date of Kuma ControlPlane version.
	BuildDate string `json:"buildDate,omitempty"`
	// True iff this Zone CP version is compatible with Global CP.
	KumaCpGlobalCompatible bool `json:"kumaCpGlobalCompatible,omitempty"`
}

// HealthCheck holds information about the received zone health check.
type HealthCheck struct {
	// Time last health check received.
	Time *Time `json:"time,omitempty"`
}

// Time serializes exactly as protobuf's well-known Timestamp does: RFC 3339 with
// nanosecond precision, always in UTC. A bare time.Time marshals with the control
// plane's local offset instead, so a control plane outside UTC would write bytes an
// older control plane parses to a different wall clock reading.
// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Format=date-time
// +kubebuilder:object:generate=false
type Time struct {
	Time time.Time `json:"-"`
}
