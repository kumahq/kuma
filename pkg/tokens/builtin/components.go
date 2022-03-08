package builtin

import (
	"github.com/kumahq/kuma/pkg/config/core"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
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

func NewDataplaneTokenValidator(resManager manager.ResourceManager, storeType store_config.StoreType) issuer.Validator {
	return issuer.NewValidator(func(meshName string) tokens.Validator {
		return tokens.NewValidator(
			tokens.NewMeshedSigningKeyAccessor(resManager, issuer.DataplaneTokenSigningKeyPrefix(meshName), meshName),
			tokens.NewRevocations(resManager, issuer.DataplaneTokenRevocationsSecretKey(meshName)),
			storeType,
		)
	})
}

func NewZoneIngressTokenValidator(resManager manager.ResourceManager, storeType store_config.StoreType) zoneingress.Validator {
	return zoneingress.NewValidator(
		tokens.NewValidator(
			tokens.NewSigningKeyAccessor(resManager, zoneingress.ZoneIngressSigningKeyPrefix),
			tokens.NewRevocations(resManager, zoneingress.ZoneIngressTokenRevocationsGlobalSecretKey),
			storeType,
		),
	)
}

func NewZoneTokenValidator(resManager manager.ResourceManager, mode core.CpMode, storeType store_config.StoreType) zone.Validator {
	var signingKeyAccessor tokens.SigningKeyAccessor

	if mode == core.Zone {
		signingKeyAccessor = tokens.NewSigningKeyFromPublicKeyAccessor(resManager, zone.SigningPublicKeyPrefix)
	} else {
		signingKeyAccessor = tokens.NewSigningKeyAccessor(resManager, zone.SigningKeyPrefix)
	}

	return zone.NewValidator(
		tokens.NewValidator(
			signingKeyAccessor,
			tokens.NewRevocations(resManager, zone.TokenRevocationsGlobalSecretKey),
			storeType,
		),
	)
}
