package types

type BootstrapVersion string

const (
	BootstrapV3 BootstrapVersion = "3"
)

// Bootstrap is sent to a client (Kuma DP) by putting YAML into a response body.
// This YAML has no information about Bootstrap version therefore we put extra header with a version
// Value of this header is then used in CLI arg --bootstrap-version when Envoy is run
const BootstrapVersionHeader = "kuma-bootstrap-version"

type BootstrapRequest struct {
	Mesh              string  `json:"mesh"`
	Name              string  `json:"name"`
	ProxyType         string  `json:"proxyType"`
	AdminPort         uint32  `json:"adminPort,omitempty"`
	DataplaneToken    string  `json:"dataplaneToken,omitempty"`
	DataplaneResource string  `json:"dataplaneResource,omitempty"`
	Host              string  `json:"-"`
	Version           Version `json:"version"`
	// CaCert is a PEM-encoded CA cert that DP uses to verify CP
	CaCert          string            `json:"caCert"`
	DynamicMetadata map[string]string `json:"dynamicMetadata"`
	DNSPort         uint32            `json:"dnsPort,omitempty"`
	EmptyDNSPort    uint32            `json:"emptyDnsPort,omitempty"`
}

type Version struct {
	KumaDp KumaDpVersion `json:"kumaDp"`
	Envoy  EnvoyVersion  `json:"envoy"`
}

type KumaDpVersion struct {
	Version   string `json:"version"`
	GitTag    string `json:"gitTag"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
}

type EnvoyVersion struct {
	Version          string `json:"version"`
	Build            string `json:"build"`
	KumaDpCompatible bool   `json:"kumaDpCompatible"`
}
