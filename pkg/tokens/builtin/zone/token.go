package zone

import (
	"github.com/golang-jwt/jwt/v4"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

const (
	SigningKeyPrefix       = "zone-token-signing-key"
	SigningPublicKeyPrefix = "zone-token-signing-public-key"
)

var TokenRevocationsGlobalSecretKey = core_model.ResourceKey{
	Name: "zone-token-revocations",
	Mesh: core_model.NoMesh,
}

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
