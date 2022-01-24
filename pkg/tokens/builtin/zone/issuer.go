package zone

import (
	"context"
	"time"

	"github.com/thediveo/enumflag"

	"github.com/kumahq/kuma/pkg/core/tokens"
)

type Token = string

type Identity struct {
	Zone  string
	Scope Scope
}

// TokenIssuer issues Zone Tokens used then for proving identity of the zone egresses.
// Issued token can be bound by zone name.
// See pkg/sds/auth/universal/authenticator.go to check algorithm for authentication
type TokenIssuer interface {
	Generate(ctx context.Context, identity Identity, validFor time.Duration) (tokens.Token, error)
}

var _ TokenIssuer = &jwtTokenIssuer{}

func NewTokenIssuer(issuer tokens.Issuer) TokenIssuer {
	return &jwtTokenIssuer{
		issuer: issuer,
	}
}

type jwtTokenIssuer struct {
	issuer tokens.Issuer
}

func (j *jwtTokenIssuer) Generate(ctx context.Context, identity Identity, validFor time.Duration) (Token, error) {
	claims := &zoneClaims{
		Zone:  identity.Zone,
		Scope: identity.Scope,
	}

	return j.issuer.Generate(ctx, claims, validFor)
}

type ScopeItem enumflag.Flag

const (
	// TODO (bartsmykla): remove comment bellow when Zone Token will be available for
	//  dataplanes and ingresses
	//
	// Zone Token can currently be used to prove Egress identity only, but I want
	// the Dataplane to be representend by int(0), Ingress as int(1) and Egress as int(2)
	// so I'm leaving them all here
	Dataplane ScopeItem = iota
	Ingress
	Egress
)

var ScopeItemsIds = map[ScopeItem][]string{
	// TODO (bartsmykla): uncomment when Zone Token will be available for dataplanes
	// 	and ingresses
	// Dataplane: {"dataplane"},
	// Ingress:   {"ingress"},
	Egress: {"egress"},
}

type Scope []ScopeItem

func (s *Scope) ContainsAtLeastOneOf(items ...ScopeItem) bool {
	for _, item := range items {
		if s.contains(item) {
			return true
		}
	}

	return false
}

func (s Scope) contains(item ScopeItem) bool {
	for _, a := range s {
		if a == item {
			return true
		}
	}

	return false
}

var FullScope Scope = []ScopeItem{
	// TODO (bartsmykla): uncomment when Zone Token will be available for dataplanes
	// 	and ingresses
	// Dataplane,
	// Ingress,
	Egress,
}
