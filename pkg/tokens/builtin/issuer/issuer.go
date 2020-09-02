package issuer

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/sds/auth"
)

type DataplaneIdentity struct {
	Name string
	Mesh string
	Tags mesh_proto.MultiValueTagSet
}

// DataplaneTokenIssuer issues Dataplane Tokens used then for proving identity of the dataplanes.
// Issued token can be bound by name, mesh or tags so you can pick your level of security.
// See pkg/sds/auth/universal/authenticator.go to check algorithm for authentication
type DataplaneTokenIssuer interface {
	Generate(identity DataplaneIdentity) (auth.Credential, error)
	Validate(credential auth.Credential) (DataplaneIdentity, error)
}

type claims struct {
	Name string
	Mesh string
	Tags map[string][]string
	jwt.StandardClaims
}

func NewDataplaneTokenIssuer(privateKey []byte) DataplaneTokenIssuer {
	return &jwtTokenIssuer{privateKey}
}

var _ DataplaneTokenIssuer = &jwtTokenIssuer{}

type jwtTokenIssuer struct {
	privateKey []byte
}

func (i *jwtTokenIssuer) Generate(identity DataplaneIdentity) (auth.Credential, error) {
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
	tokenString, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return auth.Credential(tokenString), nil
}

func (i *jwtTokenIssuer) Validate(credential auth.Credential) (DataplaneIdentity, error) {
	c := &claims{}

	token, err := jwt.ParseWithClaims(string(credential), c, func(*jwt.Token) (interface{}, error) {
		return i.privateKey, nil
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
