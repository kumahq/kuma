package issuer

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/kumahq/kuma/pkg/core/tokens"
)

// DataplaneTokenIssuer issues Dataplane Tokens used then for proving identity of the dataplanes.
// Issued token can be bound by name, mesh or tags so you can pick your level of security.
type DataplaneTokenIssuer interface {
	Generate(ctx context.Context, identity DataplaneIdentity, validFor time.Duration) (tokens.Token, error)
}

func NewDataplaneTokenIssuer(issuers func(string) tokens.Issuer) DataplaneTokenIssuer {
	return &jwtTokenIssuer{
		issuers: issuers,
	}
}

var _ DataplaneTokenIssuer = &jwtTokenIssuer{}

type jwtTokenIssuer struct {
	issuers func(string) tokens.Issuer
}

func (i *jwtTokenIssuer) Generate(ctx context.Context, identity DataplaneIdentity, validFor time.Duration) (tokens.Token, error) {
	tags := map[string][]string{}
	for tagName := range identity.Tags {
		tags[tagName] = identity.Tags.Values(tagName)
	}

	claims := &DataplaneClaims{
		Name:             identity.Name,
		Mesh:             identity.Mesh,
		Tags:             tags,
		Type:             string(identity.Type),
		RegisteredClaims: jwt.RegisteredClaims{},
	}

	return i.issuers(identity.Mesh).Generate(ctx, claims, validFor)
}
