package defaults

import (
	"context"

	"github.com/pkg/errors"

	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

func (d *defaultsComponent) createZoneIngressSigningKeyIfNotExist(ctx context.Context) error {
	signingKey, err := zoneingress.CreateSigningKey()
	if err != nil {
		return errors.Wrap(err, "could not create a Zone Ingress signing key")
	}
	key := zoneingress.SigningKeyResourceKey()
	err = d.resManager.Get(ctx, signingKey, core_store.GetBy(key))
	if err == nil {
		log.V(1).Info("Zone Ingress signing key already exists. Skip creating zone ingress signing key.")
		return nil
	}
	if !core_store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource")
	}
	log.Info("trying to create a Zone Ingress signing key")
	if err := d.resManager.Create(ctx, signingKey, core_store.CreateBy(key)); err != nil {
		log.V(1).Info("could not create a Zone Ingress signing key", "err", err)
		return errors.Wrap(err, "could not create a resource")
	}
	log.Info("Zone Ingress signing key created")
	return nil
}
