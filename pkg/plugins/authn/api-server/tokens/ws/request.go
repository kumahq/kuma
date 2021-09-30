package ws

type UserTokenRequest struct {
	Name     string `json:"name"`
	Group    string `json:"group"`
	ValidFor string `json:"validFor"`
}
