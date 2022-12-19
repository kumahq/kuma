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

type ZoneIngressClaims struct {
	Zone string
	jwt.RegisteredClaims
}

var _ core_tokens.Claims = &ZoneIngressClaims{}

func (c *ZoneIngressClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *ZoneIngressClaims) KeyIDFallback() {
}

func (c *ZoneIngressClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
