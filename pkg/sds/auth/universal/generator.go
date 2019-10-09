package universal

import (
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type claims struct {
	Name string
	Mesh string
	jwt.StandardClaims
}

func NewCredentialGenerator(PrivateKey []byte) auth.CredentialGenerator {
	return &jwtTokenGenerator{PrivateKey}
}

var _ auth.CredentialGenerator = &jwtTokenGenerator{}

type jwtTokenGenerator struct {
	privateKey []byte
}

func (i *jwtTokenGenerator) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
	c := claims{
		Name:           proxyId.Name,
		Mesh:           proxyId.Mesh,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return auth.Credential(tokenString), nil
}
