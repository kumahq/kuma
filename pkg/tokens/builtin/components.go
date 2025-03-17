package builtin

import (
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
)

var log = core.Log.WithName("tokens-validator")

func NewDataplaneTokenIssuer(resManager manager.ResourceManager) issuer.DataplaneTokenIssuer {
	return issuer.NewDataplaneTokenIssuer(func(meshName string) tokens.Issuer {
		return tokens.NewTokenIssuer(
			tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey(meshName), meshName),
		)
	})
}

func NewZoneTokenIssuer(resManager manager.ResourceManager) zone.TokenIssuer {
	return zone.NewTokenIssuer(
		tokens.NewTokenIssuer(
			tokens.NewSigningKeyManager(resManager, system.ZoneTokenSigningKeyPrefix),
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
		// Also use keys that are not bound to any mesh
		keys = append(keys, keysByMesh[""]...)

		staticSigningKeyAccessor, err := tokens.NewStaticSigningKeyAccessor(keys)
		if err != nil {
			return nil, err
		}
		accessors := []tokens.SigningKeyAccessor{staticSigningKeyAccessor}
		if cfg.UseSecrets {
			accessors = append(accessors, tokens.NewMeshedSigningKeyAccessor(resManager, system.DataplaneTokenSigningKey(meshName), meshName))
		}
		return tokens.NewValidator(
			log.WithName("dataplane-token"),
			accessors,
			tokens.NewRevocations(resManager, model.ResourceKey{Name: system.DataplaneTokenRevocations(meshName)}),
			storeType,
		), nil
	}), nil
}

func NewZoneTokenValidator(resManager manager.ReadOnlyResourceManager, isFederatedZone bool, storeType store_config.StoreType, cfg dp_server.ZoneTokenValidatorConfig) (zone.Validator, error) {
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
		if isFederatedZone {
			accessors = append(accessors, tokens.NewSigningKeyFromPublicKeyAccessor(resManager, system.ZoneTokenSigningPublicKeyPrefix))
		} else {
			accessors = append(accessors, tokens.NewSigningKeyAccessor(resManager, system.ZoneTokenSigningKeyPrefix))
		}
	}

	return zone.NewValidator(
		tokens.NewValidator(
			log.WithName("zone-token"),
			accessors,
			tokens.NewRevocations(resManager, model.ResourceKey{Name: system.ZoneTokenRevocations}),
			storeType,
		),
	), nil
}
