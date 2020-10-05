package issuer

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/sds/auth"
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

func (i *jwtTokenIssuer) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
	signingKey, err := i.signingKey()
	if err != nil {
		return "", err
	}

	c := claims{
		Name:           proxyId.Name,
		Mesh:           proxyId.Mesh,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return auth.Credential(tokenString), nil
}

func (i *jwtTokenIssuer) Validate(credential auth.Credential) (xds.ProxyId, error) {
	signingKey, err := i.signingKey()
	if err != nil {
		return xds.ProxyId{}, err
	}

	c := &claims{}

	token, err := jwt.ParseWithClaims(string(credential), c, func(*jwt.Token) (interface{}, error) {
		return signingKey, nil
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
