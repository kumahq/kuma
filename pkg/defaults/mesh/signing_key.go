package mesh

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/tokens"
)

func ensureDataplaneTokenSigningKey(ctx context.Context, resManager manager.ResourceManager, mesh model.Resource) (bool, error) {
	meshName := mesh.GetMeta().GetName()
	signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey(meshName), meshName)
	_, _, err := signingKeyManager.GetLatestSigningKey(ctx)
	if err == nil {
		return false, nil
	}
	if err != nil && !tokens.IsSigningKeyNotFound(err) {
		return false, err
	}
	if err := signingKeyManager.CreateDefaultSigningKey(ctx); err != nil {
		return false, err
	}
	return true, nil
}
