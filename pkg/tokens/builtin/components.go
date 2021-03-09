package builtin

import (
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func NewDataplaneTokenIssuer(resManager manager.ReadOnlyResourceManager) (issuer.DataplaneTokenIssuer, error) {
	return issuer.NewDataplaneTokenIssuer(func(meshName string) ([]byte, error) {
		return issuer.GetSigningKey(resManager, issuer.DataplaneTokenPrefix, meshName)
	}), nil
}
