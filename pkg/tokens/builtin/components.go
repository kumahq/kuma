package builtin

import (
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

func NewDataplaneTokenIssuer(resManager manager.ResourceManager) issuer.DataplaneTokenIssuer {
	return issuer.NewDataplaneTokenIssuer(func(meshName string) tokens.Issuer {
		return tokens.NewTokenIssuer(
			tokens.NewMeshedSigningKeyManager(resManager, issuer.DataplaneTokenSigningKeyPrefix(meshName), meshName),
		)
	})
}

func NewZoneIngressTokenIssuer(resManager manager.ResourceManager) zoneingress.TokenIssuer {
	return zoneingress.NewTokenIssuer(
		tokens.NewTokenIssuer(
			tokens.NewSigningKeyManager(resManager, zoneingress.ZoneIngressSigningKeyPrefix),
		),
	)
}

func NewZoneTokenIssuer(resManager manager.ResourceManager) zone.TokenIssuer {
	return zone.NewTokenIssuer(
		tokens.NewTokenIssuer(
			tokens.NewSigningKeyManager(resManager, zone.SigningKeyPrefix),
		),
	)
}

func NewDataplaneTokenValidator(resManager manager.ResourceManager) issuer.Validator {
	return issuer.NewValidator(func(meshName string) tokens.Validator {
		return tokens.NewValidator(
			tokens.NewMeshedSigningKeyAccessor(resManager, issuer.DataplaneTokenSigningKeyPrefix(meshName), meshName),
			tokens.NewRevocations(resManager, issuer.DataplaneTokenRevocationsSecretKey(meshName)),
		)
	})
}

func NewZoneIngressTokenValidator(resManager manager.ResourceManager) zoneingress.Validator {
	return zoneingress.NewValidator(
		tokens.NewValidator(
			tokens.NewSigningKeyAccessor(resManager, zoneingress.ZoneIngressSigningKeyPrefix),
			tokens.NewRevocations(resManager, zoneingress.ZoneIngressTokenRevocationsGlobalSecretKey),
		),
	)
}

func NewZoneTokenValidator(resManager manager.ResourceManager) zone.Validator {
	return zone.NewValidator(
		tokens.NewValidator(
			tokens.NewSigningKeyAccessor(resManager, zone.SigningKeyPrefix),
			tokens.NewRevocations(resManager, zone.TokenRevocationsGlobalSecretKey),
		),
	)
}
