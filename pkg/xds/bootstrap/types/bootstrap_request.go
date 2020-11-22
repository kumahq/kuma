package types

type BootstrapRequest struct {
	Mesh               string `json:"mesh"`
	Name               string `json:"name"`
	AdminPort          uint32 `json:"adminPort,omitempty"`
	DataplaneTokenPath string `json:"dataplaneTokenPath,omitempty"`
	DataplaneResource  string `json:"dataplaneResource,omitempty"`
	Host               string `json:"-"`
	Version			   Version `json:"version"`
}

type Version struct {
	KumaDp 	KumaDpVersion `json:"kumaDp"`
	Envoy 	EnvoyVersion `json:"envoy"`
}

type KumaDpVersion struct {
	Version 	string `json:"version"`
	GitTag 		string `json:"gitTag"`
	GitCommit 	string `json:"gitCommit"`
	BuildDate 	string `json:"buildDate"`
}

type EnvoyVersion struct {
	Version string `json:"version"`
	Build	string `json:"build"`
}
