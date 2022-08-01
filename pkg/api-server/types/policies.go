package types

type PoliciesResponse struct {
	Policies []PolicyEntry `json:"policies"`
}

type PolicyEntry struct {
	Name        string `json:"name"`
	ReadOnly    bool   `json:"readOnly"`
	Path        string `json:"path"`
	DisplayName string `json:"displayName"`
}
