package types

type DataplaneTokenRequest struct {
	Name string              `json:"name"`
	Mesh string              `json:"mesh"`
	Tags map[string][]string `json:"tags"`
}
