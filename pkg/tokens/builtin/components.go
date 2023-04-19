package builtin

import (
	"context"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

var log = core.Log.WithName("tokens-validator")

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
			tokens.NewSigningKeyManager(context.Background(), resManager, zoneingress.ZoneIngressSigningKeyPrefix),
		),
	)
}

func NewZoneTokenIssuer(resManager manager.ResourceManager) zone.TokenIssuer {
	return zone.NewTokenIssuer(
		tokens.NewTokenIssuer(
			tokens.NewSigningKeyManager(context.Background(), resManager, zone.SigningKeyPrefix),
		),
	)
}

func NewDataplaneTokenValidator(resManager manager.ReadOnlyResourceManager, storeType store_config.StoreType, cfg dp_server.DpTokenValidatorConfig) (issuer.Validator, error) {
	keysByMesh, err := tokens.PublicKeyByMeshFromConfig(cfg.PublicKeys)
	if err != nil {
		return nil, err
	}
	return issuer.NewValidator(func(meshName string) (tokens.Validator, error) {
		keys := keysByMesh[meshName]
		staticSigningKeyAccessor, err := tokens.NewStaticSigningKeyAccessor(keys)
		if err != nil {
			return nil, err
		}
		accessors := []tokens.SigningKeyAccessor{staticSigningKeyAccessor}
		if cfg.UseSecrets {
			accessors = append(accessors, tokens.NewMeshedSigningKeyAccessor(resManager, issuer.DataplaneTokenSigningKeyPrefix(meshName), meshName))
		}
		return tokens.NewValidator(
			log.WithName("dataplane-token"),
			accessors,
			tokens.NewRevocations(resManager, issuer.DataplaneTokenRevocationsSecretKey(meshName)),
			storeType,
		), nil
	}), nil
}

func NewZoneIngressTokenValidator(resManager manager.ReadOnlyResourceManager, storeType store_config.StoreType) zoneingress.Validator {
	return zoneingress.NewValidator(
		tokens.NewValidator(
			log.WithName("zone-ingress-token"),
			[]tokens.SigningKeyAccessor{tokens.NewSigningKeyAccessor(resManager, zoneingress.ZoneIngressSigningKeyPrefix)},
			tokens.NewRevocations(resManager, zoneingress.ZoneIngressTokenRevocationsGlobalSecretKey),
			storeType,
		),
	)
}

func NewZoneTokenValidator(resManager manager.ReadOnlyResourceManager, mode config_core.CpMode, storeType store_config.StoreType, cfg dp_server.ZoneTokenValidatorConfig) (zone.Validator, error) {
	publicKeys, err := tokens.PublicKeyFromConfig(cfg.PublicKeys)
	if err != nil {
		return nil, err
	}
	staticSigningKeyAccessor, err := tokens.NewStaticSigningKeyAccessor(publicKeys)
	if err != nil {
		return nil, err
	}
	accessors := []tokens.SigningKeyAccessor{staticSigningKeyAccessor}
	if cfg.UseSecrets {
		if mode == config_core.Zone {
			accessors = append(accessors, tokens.NewSigningKeyFromPublicKeyAccessor(resManager, zone.SigningPublicKeyPrefix))
		} else {
			accessors = append(accessors, tokens.NewSigningKeyAccessor(resManager, zone.SigningKeyPrefix))
		}
	}

	return zone.NewValidator(
		tokens.NewValidator(
			log.WithName("zone-token"),
			accessors,
			tokens.NewRevocations(resManager, zone.TokenRevocationsGlobalSecretKey),
			storeType,
		),
	), nil
}
