package zoneingress

import (
	"context"

	"github.com/golang-jwt/jwt/v4"

	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
)

type zoneValidatorAdapter struct {
	zoneIngressValidator Validator
	zoneValidator        zone.Validator
}

var _ zone.Validator = &zoneValidatorAdapter{}

// NewZoneValidatorAdapter returns Zone Token Validator that has a fallback on Zone Ingress Validator
// This is used for backwards compatibility to still support old ingress token.
// This should be deleted if we delete zone ingress token.
func NewZoneValidatorAdapter(zoneIngressValidator Validator, zoneValidator zone.Validator) zone.Validator {
	return &zoneValidatorAdapter{
		zoneIngressValidator: zoneIngressValidator,
		zoneValidator:        zoneValidator,
	}
}

func (z *zoneValidatorAdapter) Validate(ctx context.Context, token zone.Token) (zone.Identity, error) {
	if isZoneToken(token) {
		return z.zoneValidator.Validate(ctx, token)
	}
	id, err := z.zoneIngressValidator.Validate(ctx, token)
	if err != nil {
		return zone.Identity{}, err
	}
	return zone.Identity{
		Zone:  id.Zone,
		Scope: []string{zone.IngressScope},
	}, nil
}

func isZoneToken(token zone.Token) bool {
	parser := jwt.Parser{}
	claims := zone.ZoneClaims{}
	_, _, err := parser.ParseUnverified(token, &claims)
	if err != nil {
		return false
	}
	return len(claims.Scope) > 0 // zone token has to contain Scope
}
