package types

type ZoneIngressTokenRequest struct {
	Zone     string `json:"zone"`
	ValidFor string `json:"validFor"`
}
