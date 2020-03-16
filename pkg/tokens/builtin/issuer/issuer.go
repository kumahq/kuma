package issuer

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
)

type DataplaneTokenIssuer interface {
	Generate(proxyId xds.ProxyId) (auth.Credential, error)
	Validate(credential auth.Credential) (xds.ProxyId, error)
}

type claims struct {
	Name string
	Mesh string
	jwt.StandardClaims
}

func NewDataplaneTokenIssuer(privateKey []byte) DataplaneTokenIssuer {
	return &jwtTokenIssuer{privateKey}
}

var _ DataplaneTokenIssuer = &jwtTokenIssuer{}

type jwtTokenIssuer struct {
	privateKey []byte
}

func (i *jwtTokenIssuer) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
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

func (i *jwtTokenIssuer) Validate(credential auth.Credential) (xds.ProxyId, error) {
	c := &claims{}

	token, err := jwt.ParseWithClaims(string(credential), c, func(*jwt.Token) (interface{}, error) {
		return i.privateKey, nil
	})
	if err != nil {
		return xds.ProxyId{}, errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return xds.ProxyId{}, errors.New("token is not valid")
	}

	id := xds.ProxyId{
		Mesh: c.Mesh,
		Name: c.Name,
	}
	return id, nil
}
