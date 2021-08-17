package types

type IndexResponse struct {
	Hostname   string `json:"hostname"`
	Tagline    string `json:"tagline"`
	Version    string `json:"version"`
	InstanceId string `json:"instance_id"`
	ClusterId  string `json:"cluster_id"`
}
