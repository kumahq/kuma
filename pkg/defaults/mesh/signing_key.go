package mesh

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func ensureSigningKey(resManager manager.ResourceManager, meshName string) (err error, created bool) {
	signingKey, err := issuer.CreateSigningKey()
	if err != nil {
		return errors.Wrap(err, "could not create a signing key"), false
	}
	key := issuer.SigningKeyResourceKey(meshName)
	err = resManager.Get(context.Background(), &signingKey, core_store.GetBy(key))
	if err == nil {
		return nil, false
	}
	if !core_store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(context.Background(), &signingKey, core_store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
