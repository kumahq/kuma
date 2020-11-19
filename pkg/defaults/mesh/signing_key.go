package mesh

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func createSigningKey(resManager manager.ResourceManager, meshName string) error {
	key, err := issuer.CreateSigningKey()
	if err != nil {
		return err
	}
	if err := resManager.Create(context.Background(), &key, core_store.CreateBy(issuer.SigningKeyResourceKey(meshName))); err != nil {
		return err
	}
	return nil
}
