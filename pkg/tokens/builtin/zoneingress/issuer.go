package zoneingress

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Token = string

type Identity struct {
	Zone string
}

// TokenIssuer issues Zone Ingress Tokens used then for proving identity of the zone ingresses.
// Issued token can be bound by zone name.
// See pkg/sds/auth/universal/authenticator.go to check algorithm for authentication
type TokenIssuer interface {
	Generate(identity Identity) (Token, error)
	Validate(token Token) (Identity, error)
}

type claims struct {
	Zone string
	jwt.StandardClaims
}

type SigningKeyAccessor func() ([]byte, error)

var _ TokenIssuer = &jwtTokenIssuer{}

func NewTokenIssuer(signingKeyAccessor SigningKeyAccessor) TokenIssuer {
	return &jwtTokenIssuer{signingKeyAccessor}
}

type jwtTokenIssuer struct {
	signingKeyAccessor SigningKeyAccessor
}

func (j *jwtTokenIssuer) signingKey() ([]byte, error) {
	signingKey, err := j.signingKeyAccessor()
	if err != nil {
		return nil, err
	}
	if len(signingKey) == 0 {
		return nil, SigningKeyNotFound()
	}
	return signingKey, nil
}

func (j *jwtTokenIssuer) Generate(identity Identity) (Token, error) {
	signingKey, err := j.signingKey()
	if err != nil {
		return "", err
	}

	c := claims{
		Zone:           identity.Zone,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return tokenString, nil
}

func (j *jwtTokenIssuer) Validate(rawToken Token) (Identity, error) {
	signingKey, err := j.signingKey()
	if err != nil {
		return Identity{}, err
	}

	c := &claims{}

	token, err := jwt.ParseWithClaims(rawToken, c, func(*jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return Identity{}, errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return Identity{}, errors.New("token is not valid")
	}

	id := Identity{
		Zone: c.Zone,
	}
	return id, nil
}
