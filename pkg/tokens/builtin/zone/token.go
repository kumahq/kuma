package zone

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

const SigningKeyPrefix = "zone-token-signing-key"
const SigningPublicKeyPrefix = "zone-token-signing-public-key"

var TokenRevocationsGlobalSecretKey = core_model.ResourceKey{
	Name: "zone-token-revocations",
	Mesh: core_model.NoMesh,
}

type zoneClaims struct {
	Zone  string
	Scope []string
	jwt.RegisteredClaims
}

var _ core_tokens.Claims = &zoneClaims{}

func (c *zoneClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *zoneClaims) KeyIDFallback() (int, error) {
	return 0, errors.New("missing Key ID")
}

func (c *zoneClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
