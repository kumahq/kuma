package mesh

import (
	"context"

	tokens_builtin "github.com/kumahq/kuma/pkg/tokens/builtin"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func ensureSigningKey(resManager manager.ResourceManager, meshName string) (created bool, err error) {
	signingKey, err := issuer.CreateSigningKey()
	if err != nil {
		return false, errors.Wrap(err, "could not create a signing key")
	}
	key := issuer.SigningKeyResourceKey(tokens_builtin.DataplaneTokenPrefix, meshName)
	err = resManager.Get(context.Background(), signingKey, core_store.GetBy(key))
	if err == nil {
		return false, nil
	}
	if !core_store.IsResourceNotFound(err) {
		return false, errors.Wrap(err, "could not retrieve a resource")
	}
	if err := resManager.Create(context.Background(), signingKey, core_store.CreateBy(key)); err != nil {
		return false, errors.Wrap(err, "could not create a resource")
	}
	return true, nil
}
