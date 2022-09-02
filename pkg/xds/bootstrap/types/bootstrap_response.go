package types

type BootstrapVersion string

const (
	BootstrapV3 BootstrapVersion = "3"
)

// Bootstrap is sent to a client (Kuma DP) by putting YAML into a response body.
// This YAML has no information about Bootstrap version therefore we put extra header with a version
// Value of this header is then used in CLI arg --bootstrap-version when Envoy is run
const BootstrapVersionHeader = "kuma-bootstrap-version"

type BootstrapResponse struct {
	Bootstrap                []byte                   `json:"bootstrap"`
	KumaSidecarConfiguration KumaSidecarConfiguration `json:"kumaSidecarConfiguration"`
}

type KumaSidecarConfiguration struct {
	Networking NetworkingConfiguration `json:"networking"`
	Metrics    MetricsConfiguration    `json:"metrics"`
}

type NetworkingConfiguration struct {
	IsUsingTransparentProxy bool `json:"isUsingTransparentProxy"`
}

type MetricsConfiguration struct {
	Aggregate []Aggregate `json:"aggregate"`
}

type Aggregate struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    uint32 `json:"port"`
	Path    string `json:"path"`
}
