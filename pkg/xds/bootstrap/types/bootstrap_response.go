package types

type BootstrapResponse struct {
	Bootstrap                []byte                   `json:"bootstrap"`
	KumaSidecarConfiguration KumaSidecarConfiguration `json:"kumaSidecarConfiguration"`
}

type KumaSidecarConfiguration struct {
	Networking NetworkingConfiguration `json:"networking"`
	Metrics    MetricsConfiguration    `json:"metrics"`
}

type NetworkingConfiguration struct {
	Address string `json:"address"`
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
