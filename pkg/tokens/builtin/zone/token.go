package zone

import (
	"github.com/golang-jwt/jwt/v5"

	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

type ZoneClaims struct {
	Zone  string
	Scope []string
	jwt.RegisteredClaims
}

var _ core_tokens.Claims = &ZoneClaims{}

func (c *ZoneClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *ZoneClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
