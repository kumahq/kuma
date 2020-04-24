package issuer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	core_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
)

const defaultRsaBits = 2048

var signingKeyResourceKey = model.ResourceKey{
	Mesh: "default",
	Name: "dataplane-token-signing-key",
}

func CreateDefaultSigningKey(manager core_manager.SecretManager) error {
	key, err := createSigningKey()
	if err != nil {
		return err
	}
	return storeKeyIfNotExist(manager, key)
}

func storeKeyIfNotExist(manager core_manager.SecretManager, keyResource system.SecretResource) error {
	ctx := context.Background()
	resource := system.SecretResource{}
	if err := manager.Get(ctx, &resource, store.GetBy(signingKeyResourceKey)); err != nil {
		if store.IsResourceNotFound(err) {
			if err := manager.Create(ctx, &keyResource, store.CreateBy(signingKeyResourceKey)); err != nil {
				return errors.Wrap(err, "could not store a private key")
			}
		} else {
			return errors.Wrap(err, "could not check if private key exists")
		}
	}
	return nil
}

func createSigningKey() (system.SecretResource, error) {
	res := system.SecretResource{}
	key, err := rsa.GenerateKey(rand.Reader, defaultRsaBits)
	if err != nil {
		return res, errors.Wrap(err, "failed to generate rsa key")
	}
	res.Spec = system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: x509.MarshalPKCS1PrivateKey(key),
		},
	}
	return res, nil
}

func GetSigningKey(manager core_manager.SecretManager) ([]byte, error) {
	resource := system.SecretResource{}
	if err := manager.Get(context.Background(), &resource, store.GetBy(signingKeyResourceKey)); err != nil {
		return nil, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return resource.Spec.GetData().GetValue(), nil
}
