package ws

type UserTokenRequest struct {
	Name     string   `json:"name"`
	Groups   []string `json:"groups"`
	ValidFor string   `json:"validFor"`
}
