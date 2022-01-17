package types

type IndexResponse struct {
	Hostname   string `json:"hostname"`
	Tagline    string `json:"tagline"`
	Version    string `json:"version"`
	InstanceId string `json:"instanceId"`
	ClusterId  string `json:"clusterId"`
	GuiURL     string `json:"gui,omitempty"`
}
