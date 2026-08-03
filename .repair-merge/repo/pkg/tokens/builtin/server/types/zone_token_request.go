package types

type ZoneTokenRequest struct {
	Zone     string   `json:"zone"`
	Scope    []string `json:"scope"`
	ValidFor string   `json:"validFor"`
}
