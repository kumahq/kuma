package xds

type OtelSignal string

const (
	OtelSignalTraces  OtelSignal = "traces"
	OtelSignalLogs    OtelSignal = "logs"
	OtelSignalMetrics OtelSignal = "metrics"
)

const FieldOtelEnvInventory = "otelEnvInventory"

type OtelProtocol string

const (
	OtelProtocolUnknown      OtelProtocol = "unknown"
	OtelProtocolGRPC         OtelProtocol = "grpc"
	OtelProtocolHTTPProtobuf OtelProtocol = "http/protobuf"
)

type OtelAuthMode string

const (
	OtelAuthModeNone    OtelAuthMode = "none"
	OtelAuthModeHeaders OtelAuthMode = "headers"
	OtelAuthModeTLS     OtelAuthMode = "tls"
	OtelAuthModeMTLS    OtelAuthMode = "mtls"
)

type OtelBootstrapInventory struct {
	PipeEnabled      bool                    `json:"pipeEnabled,omitempty"`
	Shared           *OtelSignalEnvInventory `json:"shared,omitempty"`
	Traces           *OtelSignalEnvInventory `json:"traces,omitempty"`
	Logs             *OtelSignalEnvInventory `json:"logs,omitempty"`
	Metrics          *OtelSignalEnvInventory `json:"metrics,omitempty"`
	ValidationErrors []string                `json:"validationErrors,omitempty"`
}

type OtelSignalEnvInventory struct {
	EndpointPresent          bool         `json:"endpointPresent,omitempty"`
	EndpointParsedAsURL      bool         `json:"endpointParsedAsURL,omitempty"`
	EndpointHasPath          bool         `json:"endpointHasPath,omitempty"`
	ProtocolPresent          bool         `json:"protocolPresent,omitempty"`
	HeadersPresent           bool         `json:"headersPresent,omitempty"`
	TimeoutPresent           bool         `json:"timeoutPresent,omitempty"`
	CompressionPresent       bool         `json:"compressionPresent,omitempty"`
	InsecurePresent          bool         `json:"insecurePresent,omitempty"`
	CertificatePresent       bool         `json:"certificatePresent,omitempty"`
	ClientCertificatePresent bool         `json:"clientCertificatePresent,omitempty"`
	ClientKeyPresent         bool         `json:"clientKeyPresent,omitempty"`
	EffectiveProtocol        OtelProtocol `json:"effectiveProtocol,omitempty"`
	EffectiveAuthMode        OtelAuthMode `json:"effectiveAuthMode,omitempty"`
	OverrideKinds            []string     `json:"overrideKinds,omitempty"`
}

func (i *OtelBootstrapInventory) GetSignal(signal OtelSignal) *OtelSignalEnvInventory {
	if i == nil {
		return nil
	}

	switch signal {
	case OtelSignalTraces:
		return i.Traces
	case OtelSignalLogs:
		return i.Logs
	case OtelSignalMetrics:
		return i.Metrics
	default:
		return nil
	}
}

func (i *OtelSignalEnvInventory) HasAnyInput() bool {
	if i == nil {
		return false
	}

	return i.EndpointPresent ||
		i.ProtocolPresent ||
		i.HeadersPresent ||
		i.TimeoutPresent ||
		i.CompressionPresent ||
		i.InsecurePresent ||
		i.CertificatePresent ||
		i.ClientCertificatePresent ||
		i.ClientKeyPresent ||
		len(i.OverrideKinds) > 0 ||
		i.EffectiveProtocol != "" ||
		i.EffectiveAuthMode != ""
}
