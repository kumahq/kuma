package types

type BootstrapRequest struct {
	Mesh               string `json:"mesh"`
	Name               string `json:"name"`
	AdminPort          uint32 `json:"adminPort,omitempty"`
	DataplaneTokenPath string `json:"dataplaneTokenPath,omitempty"`
}
