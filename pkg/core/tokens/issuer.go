package tokens

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
)

// Issuer generates tokens.
// Token is a JWT token with claims that is provided by the actual issuer (for example - Dataplane Token Issuer, User Token Issuer).
// We place "kid" in token, so we don't have to validate the token against every single signing key.
// Instead, we take "kid" from the token, retrieve signing key and validate only against this key.
// A new token is always generated by using the latest signing key.
type Issuer interface {
	Generate(ctx context.Context, claims Claims, validFor time.Duration) (Token, error)
}

type jwtTokenIssuer struct {
	signingKeyManager SigningKeyManager
}

func NewTokenIssuer(signingKeyAccessor SigningKeyManager) Issuer {
	return &jwtTokenIssuer{
		signingKeyManager: signingKeyAccessor,
	}
}

var _ Issuer = &jwtTokenIssuer{}

func (j *jwtTokenIssuer) Generate(ctx context.Context, claims Claims, validFor time.Duration) (Token, error) {
	signingKey, serialNumber, err := j.signingKeyManager.GetLatestSigningKey(ctx)
	if err != nil {
		return "", err
	}

	now := core.Now()
	claims.SetRegisteredClaims(jwt.RegisteredClaims{
		ID:        core.NewUUID(),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(time.Minute * -5)), // todo(jakubdyszkiewicz) parametrize via config and go through all clock skews in the project
		ExpiresAt: jwt.NewNumericDate(now.Add(validFor)),
	})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header[KeyIDHeader] = serialNumber
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return tokenString, nil
}
