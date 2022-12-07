package types

type IndexResponse struct {
	Hostname string `json:"hostname"`
	// Deprecated (use Product instead)
	Tagline     string `json:"tagline"`
	Product     string `json:"product"`
	Version     string `json:"version"`
	InstanceId  string `json:"instanceId"`
	ClusterId   string `json:"clusterId"`
	GuiURL      string `json:"gui,omitempty"`
	BasedOnKuma string `json:"basedOnKuma,omitempty"`
}
