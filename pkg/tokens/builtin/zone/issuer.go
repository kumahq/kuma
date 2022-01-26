package zone

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core/tokens"
)

type Token = string

type Identity struct {
	Zone  string
	Scope []string
}

// TokenIssuer issues Zone Tokens used then for proving identity of the zone egresses.
// Issued token can be bound by the zone name and the scope.
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
