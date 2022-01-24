package types

import "github.com/kumahq/kuma/pkg/tokens/builtin/zone"

type ZoneTokenRequest struct {
	Zone     string           `json:"zone"`
	Scope    []zone.ScopeItem `json:"scope"`
	ValidFor string           `json:"validFor"`
}
