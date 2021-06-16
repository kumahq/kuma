package zoneingress

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

const (
	defaultRsaBits = 2048
	signingKeyName = "zone-ingress-token-signing-key"
)

func SigningKeyNotFound() error {
	return errors.Errorf("there is no Zone Ingress Signing Key in the Control Plane.")
}

func SigningKeyResourceKey() core_model.ResourceKey {
	return core_model.ResourceKey{
		Mesh: core_model.NoMesh,
		Name: signingKeyName,
	}
}

func IsSigningKeyResource(resKey core_model.ResourceKey) bool {
	return resKey.Name == signingKeyName && resKey.Mesh == core_model.NoMesh
}

func GetSigningKey(manager manager.ReadOnlyResourceManager) ([]byte, error) {
	resource := system.NewGlobalSecretResource()
	if err := manager.Get(context.Background(), resource, store.GetBy(SigningKeyResourceKey())); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, SigningKeyNotFound()
		}
		return nil, errors.Wrap(err, "could not retrieve global signing key from secret manager")
	}
	return resource.Spec.GetData().GetValue(), nil
}

func CreateSigningKey() (*system.GlobalSecretResource, error) {
	res := system.NewGlobalSecretResource()
	key, err := rsa.GenerateKey(rand.Reader, defaultRsaBits)
	if err != nil {
		return res, errors.Wrap(err, "failed to generate rsa key")
	}
	res.Spec = &system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: x509.MarshalPKCS1PrivateKey(key),
		},
	}
	return res, nil
}
