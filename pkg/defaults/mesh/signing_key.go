package mesh

import (
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func ensureDataplaneTokenSigningKey(resManager manager.ResourceManager, meshName string) (created bool, err error) {
	signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, issuer.DataplaneTokenSigningKeyPrefix(meshName), meshName)
	_, _, err = signingKeyManager.GetLatestSigningKey()
	if err == nil {
		return false, nil
	}
	if err != nil && !tokens.IsSigningKeyNotFound(err) {
		return false, err
	}
	if err := signingKeyManager.CreateDefaultSigningKey(); err != nil {
		return false, err
	}
	return true, nil
}
