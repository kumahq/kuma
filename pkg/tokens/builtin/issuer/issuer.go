package issuer

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Token = string

type DataplaneIdentity struct {
	Name string
	Mesh string
	Tags mesh_proto.MultiValueTagSet
}

// DataplaneTokenIssuer issues Dataplane Tokens used then for proving identity of the dataplanes.
// Issued token can be bound by name, mesh or tags so you can pick your level of security.
// See pkg/sds/auth/universal/authenticator.go to check algorithm for authentication
type DataplaneTokenIssuer interface {
	Generate(identity DataplaneIdentity) (Token, error)
	Validate(token Token) (DataplaneIdentity, error)
}

type claims struct {
	Name string
	Mesh string
	Tags map[string][]string
	jwt.StandardClaims
}

type SigningKeyAccessor func() ([]byte, error)

func NewDataplaneTokenIssuer(signingKeyAccessor SigningKeyAccessor) DataplaneTokenIssuer {
	return &jwtTokenIssuer{signingKeyAccessor}
}

var _ DataplaneTokenIssuer = &jwtTokenIssuer{}

type jwtTokenIssuer struct {
	signingKeyAccessor SigningKeyAccessor
}

func (i *jwtTokenIssuer) signingKey() ([]byte, error) {
	signingKey, err := i.signingKeyAccessor()
	if err != nil {
		return nil, err
	}
	if len(signingKey) == 0 {
		return nil, SigningKeyNotFound
	}
	return signingKey, nil
}

func (i *jwtTokenIssuer) Generate(identity DataplaneIdentity) (Token, error) {
	signingKey, err := i.signingKey()
	if err != nil {
		return "", err
	}

	tags := map[string][]string{}
	for tagName := range identity.Tags {
		tags[tagName] = identity.Tags.Values(tagName)
	}

	c := claims{
		Name:           identity.Name,
		Mesh:           identity.Mesh,
		Tags:           tags,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return tokenString, nil
}

func (i *jwtTokenIssuer) Validate(rawToken Token) (DataplaneIdentity, error) {
	signingKey, err := i.signingKey()
	if err != nil {
		return DataplaneIdentity{}, err
	}

	c := &claims{}

	token, err := jwt.ParseWithClaims(rawToken, c, func(*jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return DataplaneIdentity{}, errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return DataplaneIdentity{}, errors.New("token is not valid")
	}

	id := DataplaneIdentity{
		Mesh: c.Mesh,
		Name: c.Name,
		Tags: mesh_proto.MultiValueTagSetFrom(c.Tags),
	}
	return id, nil
}
