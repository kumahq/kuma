package zoneegress

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

const ZoneEgressSigningKeyPrefix = "zone-egress-token-signing-key"

var ZoneEgressTokenRevocationsGlobalSecretKey = core_model.ResourceKey{
	Name: "zone-egress-token-revocations",
	Mesh: core_model.NoMesh,
}

type zoneEgressClaims struct {
	Zone string
	jwt.RegisteredClaims
}

var _ core_tokens.Claims = &zoneEgressClaims{}

func (c *zoneEgressClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *zoneEgressClaims) KeyIDFallback() (int, error) {
	return 0, errors.New("missing Key ID")
}

func (c *zoneEgressClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
