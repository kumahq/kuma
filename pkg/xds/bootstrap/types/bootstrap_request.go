package types

type BootstrapRequest struct {
	Mesh               string  `json:"mesh"`
	Name               string  `json:"name"`
	ProxyType          string  `json:"proxyType"`
	DataplaneToken     string  `json:"dataplaneToken,omitempty"`
	DataplaneTokenPath string  `json:"dataplaneTokenPath,omitempty"`
	DataplaneResource  string  `json:"dataplaneResource,omitempty"`
	Host               string  `json:"-"`
	Version            Version `json:"version"`
	// CaCert is a PEM-encoded CA cert that DP uses to verify CP
<<<<<<< HEAD
	CaCert              string            `json:"caCert"`
	DynamicMetadata     map[string]string `json:"dynamicMetadata"`
	DNSPort             uint32            `json:"dnsPort,omitempty"`
	EmptyDNSPort        uint32            `json:"emptyDnsPort,omitempty"`
	OperatingSystem     string            `json:"operatingSystem"`
	Features            []string          `json:"features"`
	Resources           ProxyResources    `json:"resources"`
	Workdir             string            `json:"workdir"`
	AccessLogSocketPath string            `json:"accessLogSocketPath"`
	MetricsResources    MetricsResources  `json:"metricsResources"`
=======
	CaCert           string            `json:"caCert"`
	DynamicMetadata  map[string]string `json:"dynamicMetadata"`
	DNSPort          uint32            `json:"dnsPort,omitempty"`
	ReadinessPort    uint32            `json:"readinessPort,omitempty"`
	OperatingSystem  string            `json:"operatingSystem"`
	Features         []string          `json:"features"`
	Resources        ProxyResources    `json:"resources"`
	Workdir          string            `json:"workdir"`
	MetricsResources MetricsResources  `json:"metricsResources"`
	SystemCaPath     string            `json:"systemCaPath"`
>>>>>>> 20208eb60 (feat(kuma-dp): add a separate component to handle kuma-sidecar readiness probes (#11107))
}

type Version struct {
	KumaDp KumaDpVersion `json:"kumaDp"`
	Envoy  EnvoyVersion  `json:"envoy"`
}

type MetricsResources struct {
	SocketPath string `json:"socketPath"`
	CertPath   string `json:"certPath"`
	KeyPath    string `json:"keyPath"`
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

// ProxyResources contains information about what resources this proxy has
// available
type ProxyResources struct {
	MaxHeapSizeBytes uint64 `json:"maxHeapSizeBytes"`
}
