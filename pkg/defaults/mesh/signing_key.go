package mesh

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func ensureDataplaneTokenSigningKey(resManager manager.ResourceManager, meshName string) (created bool, err error) {
	return ensureSigningKeyForPrefix(resManager, meshName, issuer.DataplaneTokenPrefix)
}

func ensureEnvoyAdminClientSigningKey(resManager manager.ResourceManager, meshName string) (created bool, err error) {
	return ensureSigningKeyForPrefix(resManager, meshName, issuer.EnvoyAdminClientTokenPrefix)
}

func ensureSigningKeyForPrefix(resManager manager.ResourceManager, meshName, prefix string) (created bool, err error) {
	signingKey, err := issuer.CreateSigningKey()
	if err != nil {
		return false, errors.Wrap(err, "could not create a signing key")
	}
	key := issuer.SigningKeyResourceKey(prefix, meshName)
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
