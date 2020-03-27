package types

const TaglineKuma = "Kuma"

type IndexResponse struct {
	Hostname string `json:"hostname"`
	Tagline  string `json:"tagline"`
	Version  string `json:"version"`
}
