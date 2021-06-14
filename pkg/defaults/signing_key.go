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
		return errors.Wrap(err, "could not create a signing key")
	}
	key := zoneingress.SigningKeyResourceKey()
	err = d.resManager.Get(ctx, signingKey, core_store.GetBy(key))
	if err == nil {
		return nil
	}
	if !core_store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource")
	}
	if err := d.resManager.Create(ctx, signingKey, core_store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource")
	}
	return nil
}
