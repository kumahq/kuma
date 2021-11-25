package zoneingress

import (
	"github.com/golang-jwt/jwt/v4"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

const ZoneIngressSigningKeyPrefix = "zone-ingress-token-signing-key"

var ZoneIngressTokenRevocationsGlobalSecretKey = core_model.ResourceKey{
	Name: "zone-ingress-token-revocations",
	Mesh: core_model.NoMesh,
}

type zoneIngressClaims struct {
	Zone string
	jwt.RegisteredClaims
}

var _ core_tokens.Claims = &zoneIngressClaims{}

func (c *zoneIngressClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *zoneIngressClaims) KeyIDFallback() (int, error) {
	return 0, nil
}

func (c *zoneIngressClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
