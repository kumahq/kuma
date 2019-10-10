package types

const TaglineKuma = "Kuma"

type IndexResponse struct {
	Tagline string `json:"tagline"`
	Version string `json:"version"`
}
