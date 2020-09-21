package issuer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

var log = core.Log.WithName("tokens")

const defaultRsaBits = 2048

var SigningKeyNotFound = errors.New("there is no Signing Key in the Control Plane. If you run multi-zone setup, make sure Remote is connected to the Global before generating tokens.")

var SigningKeyResourceKey = model.ResourceKey{
	Mesh: "default",
	Name: "dataplane-token-signing-key",
}

func CreateSigningKey() (system.SecretResource, error) {
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
func GetSigningKey(manager manager.ReadOnlyResourceManager) ([]byte, error) {
	resource := system.SecretResource{}
	if err := manager.Get(context.Background(), &resource, store.GetBy(SigningKeyResourceKey)); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, SigningKeyNotFound
		}
		return nil, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return resource.Spec.GetData().GetValue(), nil
}
