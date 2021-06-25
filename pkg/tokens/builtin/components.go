package builtin

import (
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

func NewDataplaneTokenIssuer(resManager manager.ReadOnlyResourceManager) (issuer.DataplaneTokenIssuer, error) {
	return issuer.NewDataplaneTokenIssuer(func(meshName string) ([]byte, error) {
		return issuer.GetSigningKey(resManager, issuer.DataplaneTokenPrefix, meshName)
	}), nil
}

func NewZoneIngressTokenIssuer(resManager manager.ReadOnlyResourceManager) (zoneingress.TokenIssuer, error) {
	return zoneingress.NewTokenIssuer(func() ([]byte, error) {
		return zoneingress.GetSigningKey(resManager)
	}), nil
}
