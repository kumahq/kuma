package v1alpha1

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"
)

// DataplaneInsight defines the observed state of a Dataplane.
type DataplaneInsight struct {
	// Subscriptions created by a given Dataplane.
	Subscriptions []*DiscoverySubscription `json:"subscriptions,omitempty"`
	// MTLS carries the insights about mTLS for the Dataplane.
	MTLS *DataplaneInsight_MTLS `json:"mTLS,omitempty"`
	// Metadata is free form, so it keeps whatever keys the proxy reported even when
	// the control plane does not know them.
	Metadata *Struct `json:"metadata,omitempty"`
	// OpenTelemetry carries the insights about OTel runtime resolution.
	OpenTelemetry *DataplaneInsight_OpenTelemetry `json:"openTelemetry,omitempty"`
}

func (i *DataplaneInsight) GetSubscriptions() []*DiscoverySubscription {
	if i == nil {
		return nil
	}
	return i.Subscriptions
}

func (i *DataplaneInsight) GetMTLS() *DataplaneInsight_MTLS {
	if i == nil {
		return nil
	}
	return i.MTLS
}

func (i *DataplaneInsight) GetMetadata() *Struct {
	if i == nil {
		return nil
	}
	return i.Metadata
}

func (i *DataplaneInsight) GetOpenTelemetry() *DataplaneInsight_OpenTelemetry {
	if i == nil {
		return nil
	}
	return i.OpenTelemetry
}

// Struct carries what protobuf held in a Struct: an object whose keys are not known
// ahead of time. Protobuf wrote it as the object itself rather than a wrapper.
type Struct struct {
	Fields map[string]any
}

func NewStruct(fields map[string]any) *Struct {
	return &Struct{Fields: fields}
}

// StructFromProto keeps a metadata that was never reported absent rather than turning it
// into an empty object, which is what protobuf did with an unset Struct.
func StructFromProto(fields *structpb.Struct) *Struct {
	if fields == nil {
		return nil
	}
	return &Struct{Fields: fields.AsMap()}
}

func (s *Struct) GetFields() map[string]any {
	if s == nil {
		return nil
	}
	return s.Fields
}

// ToProto renders the metadata back as the protobuf Struct the xDS side works in.
func (s *Struct) ToProto() *structpb.Struct {
	if s == nil {
		return nil
	}
	out, err := structpb.NewStruct(s.Fields)
	if err != nil {
		return nil
	}
	return out
}

func (s Struct) MarshalJSON() ([]byte, error) {
	if s.Fields == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(s.Fields)
}

func (s *Struct) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Fields)
}

func (s *Struct) DeepCopyInto(out *Struct) {
	if s.Fields == nil {
		out.Fields = nil
		return
	}
	raw, err := json.Marshal(s.Fields)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, &out.Fields)
}

func (s *Struct) DeepCopy() *Struct {
	if s == nil {
		return nil
	}
	out := new(Struct)
	s.DeepCopyInto(out)
	return out
}

// DataplaneInsight_MTLS defines insights for mTLS.
type DataplaneInsight_MTLS struct {
	// Expiration time of the last certificate that was generated for a Dataplane.
	CertificateExpirationTime *Time `json:"certificateExpirationTime,omitempty"`
	// Time on which the last certificate was generated.
	LastCertificateRegeneration *Time `json:"lastCertificateRegeneration,omitempty"`
	// Number of certificate regenerations for a Dataplane.
	CertificateRegenerations uint32 `json:"certificateRegenerations,omitempty"`
	// Backend that was used to generate the current certificate.
	IssuedBackend string `json:"issuedBackend,omitempty"`
	// Supported backends (CA).
	SupportedBackends []string `json:"supportedBackends,omitempty"`
}

func (m *DataplaneInsight_MTLS) GetCertificateExpirationTime() *Time {
	if m == nil {
		return nil
	}
	return m.CertificateExpirationTime
}

func (m *DataplaneInsight_MTLS) GetLastCertificateRegeneration() *Time {
	if m == nil {
		return nil
	}
	return m.LastCertificateRegeneration
}

func (m *DataplaneInsight_MTLS) GetCertificateRegenerations() uint32 {
	if m == nil {
		return 0
	}
	return m.CertificateRegenerations
}

func (m *DataplaneInsight_MTLS) GetIssuedBackend() string {
	if m == nil {
		return ""
	}
	return m.IssuedBackend
}

func (m *DataplaneInsight_MTLS) GetSupportedBackends() []string {
	if m == nil {
		return nil
	}
	return m.SupportedBackends
}

// DataplaneInsight_OpenTelemetry carries the OTel runtime resolution insights.
type DataplaneInsight_OpenTelemetry struct {
	Backends []*DataplaneInsight_OpenTelemetry_Backend `json:"backends,omitempty"`
}

func (o *DataplaneInsight_OpenTelemetry) GetBackends() []*DataplaneInsight_OpenTelemetry_Backend {
	if o == nil {
		return nil
	}
	return o.Backends
}

// DataplaneInsight_OpenTelemetry_Backend is one resolved OTel backend.
type DataplaneInsight_OpenTelemetry_Backend struct {
	Name    string                                 `json:"name,omitempty"`
	Traces  *DataplaneInsight_OpenTelemetry_Signal `json:"traces,omitempty"`
	Logs    *DataplaneInsight_OpenTelemetry_Signal `json:"logs,omitempty"`
	Metrics *DataplaneInsight_OpenTelemetry_Signal `json:"metrics,omitempty"`
}

func (b *DataplaneInsight_OpenTelemetry_Backend) GetName() string {
	if b == nil {
		return ""
	}
	return b.Name
}

func (b *DataplaneInsight_OpenTelemetry_Backend) GetTraces() *DataplaneInsight_OpenTelemetry_Signal {
	if b == nil {
		return nil
	}
	return b.Traces
}

func (b *DataplaneInsight_OpenTelemetry_Backend) GetLogs() *DataplaneInsight_OpenTelemetry_Signal {
	if b == nil {
		return nil
	}
	return b.Logs
}

func (b *DataplaneInsight_OpenTelemetry_Backend) GetMetrics() *DataplaneInsight_OpenTelemetry_Signal {
	if b == nil {
		return nil
	}
	return b.Metrics
}

// DataplaneInsight_OpenTelemetry_Signal is the resolution of one OTel signal.
type DataplaneInsight_OpenTelemetry_Signal struct {
	Enabled         bool     `json:"enabled,omitempty"`
	EnvAllowed      bool     `json:"envAllowed,omitempty"`
	EnvInputPresent bool     `json:"envInputPresent,omitempty"`
	State           string   `json:"state,omitempty"`
	OverrideKinds   []string `json:"overrideKinds,omitempty"`
	MissingFields   []string `json:"missingFields,omitempty"`
	BlockedReasons  []string `json:"blockedReasons,omitempty"`
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetEnabled() bool {
	if s == nil {
		return false
	}
	return s.Enabled
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetEnvAllowed() bool {
	if s == nil {
		return false
	}
	return s.EnvAllowed
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetEnvInputPresent() bool {
	if s == nil {
		return false
	}
	return s.EnvInputPresent
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetState() string {
	if s == nil {
		return ""
	}
	return s.State
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetOverrideKinds() []string {
	if s == nil {
		return nil
	}
	return s.OverrideKinds
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetMissingFields() []string {
	if s == nil {
		return nil
	}
	return s.MissingFields
}

func (s *DataplaneInsight_OpenTelemetry_Signal) GetBlockedReasons() []string {
	if s == nil {
		return nil
	}
	return s.BlockedReasons
}
