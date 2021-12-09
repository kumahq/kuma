package zoneingress

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core/tokens"
)

type Token = string

type Identity struct {
	Zone string
}

// TokenIssuer issues Zone Ingress Tokens used then for proving identity of the zone ingresses.
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
	claims := &zoneIngressClaims{
		Zone: identity.Zone,
	}
	return j.issuer.Generate(ctx, claims, validFor)
}
