package manager

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/gogo/protobuf/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var privateKeyResourceKey = model.ResourceKey{
	Mesh:      "default",
	Namespace: "default",
	Name:      "initial-token-private-key",
}

func CreateDefaultPrivateKey(manager SecretManager) error {
	ctx := context.Background()
	resource := system.SecretResource{}
	if err := manager.Get(ctx, &resource, store.GetBy(privateKeyResourceKey)); err != nil {
		if store.IsResourceNotFound(err) {
			resource.Spec = types.BytesValue{
				Value: []byte(uuid.New().String()),
			}
			if err := manager.Create(ctx, &resource, store.CreateBy(privateKeyResourceKey)); err != nil {
				return errors.Wrap(err, "could not store a private key")
			}
		} else {
			return errors.Wrap(err, "could not check if private key exists")
		}
	}
	return nil
}

func GetPrivateKey(manager SecretManager) ([]byte, error) {
	resource := system.SecretResource{}
	if err := manager.Get(context.Background(), &resource, store.GetBy(privateKeyResourceKey)); err != nil {
		return nil, errors.Wrap(err, "could not retrieve private key from secret manager")
	}
	return resource.Spec.Value, nil
}
