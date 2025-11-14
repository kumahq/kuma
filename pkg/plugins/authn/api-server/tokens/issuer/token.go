package issuer

import (
	"github.com/golang-jwt/jwt/v4"

	"github.com/kumahq/kuma/v2/pkg/core/tokens"
	"github.com/kumahq/kuma/v2/pkg/core/user"
)

type UserClaims struct {
	user.User
	jwt.RegisteredClaims
}

var _ tokens.Claims = &UserClaims{}

func (c *UserClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *UserClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
