package types

type PoliciesResponse struct {
	Policies []PolicyEntry `json:"policies"`
}

type PolicyEntry struct {
	Name                string `json:"name"`
	ReadOnly            bool   `json:"readOnly"`
	Path                string `json:"path"`
	SingularDisplayName string `json:"singularDisplayName"`
	PluralDisplayName   string `json:"pluralDisplayName"`
	IsExperimental      bool   `json:"isExperimental"`
	IsTargetRefBased    bool   `json:"isTargetRefBased"`
	IsInbound           bool   `json:"isInbound"`
	IsOutbound          bool   `json:"isOutbound"`
}
