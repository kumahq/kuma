// +kubebuilder:object:generate=true
package v1alpha1

import "time"

// ZoneInsight defines the observed state of a Zone Kuma CP.
// EnvoyAdminStreams is deprecated, KDSStreams replaces it.
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
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	Subscriptions     []*KDSSubscription `json:"subscriptions,omitempty"`
	EnvoyAdminStreams *EnvoyAdminStreams `json:"envoyAdminStreams,omitempty"`
	HealthCheck       *HealthCheck       `json:"healthCheck,omitempty"`
	KDSStreams        *KDSStreams        `json:"kdsStreams,omitempty"`
}

type EnvoyAdminStreams struct {
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	ConfigDumpGlobalInstanceID string `json:"configDumpGlobalInstanceId,omitempty"`
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	StatsGlobalInstanceID string `json:"statsGlobalInstanceId,omitempty"`
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	ClustersGlobalInstanceID string `json:"clustersGlobalInstanceId,omitempty"`
}

type KDSStreams struct {
	Clusters     *KDSStream `json:"clusters,omitempty"`
	ConfigDump   *KDSStream `json:"configDump,omitempty"`
	Stats        *KDSStream `json:"stats,omitempty"`
	GlobalToZone *KDSStream `json:"globalToZone,omitempty"`
	ZoneToGlobal *KDSStream `json:"zoneToGlobal,omitempty"`
}

type KDSStream struct {
	// +kuma:nolint // protobuf omitted an empty value entirely, so the field keeps omitempty without being a pointer to stay byte compatible with what 2.14 wrote
	GlobalInstanceID string `json:"globalInstanceId,omitempty"`
	ConnectTime      *Time  `json:"connectTime,omitempty"`
}

// KDSSubscription describes a single KDS subscription created by a Zone to the Global.
// Ideally there is one per Zone lifecycle. Several of them indicate a transient loss of
// connectivity, or a restart of either control plane.
// Generation is bumped by the status sink rather than by the connection itself, and
// ZoneInstanceID names whichever Zone CP held leadership when the stream opened.
type KDSSubscription struct {
	ID                string                 `json:"id,omitempty"`
	GlobalInstanceID  string                 `json:"globalInstanceId,omitempty"`
	ConnectTime       *Time                  `json:"connectTime,omitempty"`
	DisconnectTime    *Time                  `json:"disconnectTime,omitempty"`
	Status            *KDSSubscriptionStatus `json:"status,omitempty"`
	Version           *Version               `json:"version,omitempty"`
	Generation        uint32                 `json:"generation,omitempty"`
	Config            string                 `json:"config,omitempty"`
	AuthTokenProvided bool                   `json:"authTokenProvided,omitempty"`
	ZoneInstanceID    string                 `json:"zoneInstanceId,omitempty"`
}

type KDSSubscriptionStatus struct {
	LastUpdateTime *Time                       `json:"lastUpdateTime,omitempty"`
	Total          *KDSServiceStats            `json:"total,omitempty"`
	Stat           map[string]*KDSServiceStats `json:"stat,omitempty"`
}

// KDSServiceStats counters serialize as JSON strings because protobuf writes 64 bit
// integers that way, and an older control plane rejects them when written as numbers.
type KDSServiceStats struct {
	ResponsesSent         uint64 `json:"responsesSent,omitempty,string"`
	ResponsesAcknowledged uint64 `json:"responsesAcknowledged,omitempty,string"`
	ResponsesRejected     uint64 `json:"responsesRejected,omitempty,string"`
}

type Version struct {
	KumaCP *KumaCpVersion `json:"kumaCp,omitempty"`
}

type KumaCpVersion struct {
	Version                string `json:"version,omitempty"`
	GitTag                 string `json:"gitTag,omitempty"`
	GitCommit              string `json:"gitCommit,omitempty"`
	BuildDate              string `json:"buildDate,omitempty"`
	KumaCpGlobalCompatible bool   `json:"kumaCpGlobalCompatible,omitempty"`
}

type HealthCheck struct {
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
