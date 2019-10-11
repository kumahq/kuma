package issuer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	core_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

const defaultRsaBits = 2048

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
			key, err := rsa.GenerateKey(rand.Reader, defaultRsaBits)
			if err != nil {
				return errors.Wrap(err, "failed to generate rsa key")
			}
			resource.Spec = types.BytesValue{
				Value: x509.MarshalPKCS1PrivateKey(key),
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
