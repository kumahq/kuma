package issuer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"strings"

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

const (
	defaultRsaBits = 2048

	DataplaneTokenPrefix        = "dataplane-token"
	EnvoyAdminClientTokenPrefix = "envoy-admin-client-token"
)

func SigningKeyNotFound(meshName string) error {
	return errors.Errorf("there is no Signing Key in the Control Plane for Mesh %q. Make sure the Mesh exist. If you run multi-zone setup, make sure Remote is connected to the Global before generating tokens.", meshName)
}

func IsSigningKeyNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "there is no Signing Key in the Control Plane for Mesh")
}

func SigningKeyResourceKey(prefix, meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("%s-signing-key-%s", prefix, meshName),
	}
}

func CreateSigningKey() (*system.SecretResource, error) {
	res := system.NewSecretResource()
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
func GetSigningKey(manager manager.ReadOnlyResourceManager, prefix, meshName string) ([]byte, error) {
	resource := system.NewSecretResource()
	if err := manager.Get(context.Background(), resource, store.GetBy(SigningKeyResourceKey(prefix, meshName))); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, SigningKeyNotFound(meshName)
		}
		return nil, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return resource.Spec.GetData().GetValue(), nil
}
