package issuer

import (
	"context"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	core_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

var signingKeyResourceKey = model.ResourceKey{
	Mesh:      "default",
	Namespace: "default",
	Name:      "dataplane-token-signing-key",
}

func CreateDefaultSigningKey(manager core_manager.SecretManager) error {
	ctx := context.Background()
	resource := system.SecretResource{}
	if err := manager.Get(ctx, &resource, store.GetBy(signingKeyResourceKey)); err != nil {
		if store.IsResourceNotFound(err) {
			resource.Spec = types.BytesValue{
				Value: []byte(core.NewUUID()),
			}
			if err := manager.Create(ctx, &resource, store.CreateBy(signingKeyResourceKey)); err != nil {
				return errors.Wrap(err, "could not store a private key")
			}
		} else {
			return errors.Wrap(err, "could not check if private key exists")
		}
	}
	return nil
}

func GetSigningKey(manager core_manager.SecretManager) ([]byte, error) {
	resource := system.SecretResource{}
	if err := manager.Get(context.Background(), &resource, store.GetBy(signingKeyResourceKey)); err != nil {
		return nil, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return resource.Spec.Value, nil
}
