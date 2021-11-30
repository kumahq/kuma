package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

// NewMeshedSigningKeyManager builds SigningKeyManager that is bound to a Mesh.
// Some tokens like Dataplane Token are bound to a mesh.
// In this case, singing key is also stored as a Secret in the Mesh, not as GlobalSecret.
func NewMeshedSigningKeyManager(manager manager.ResourceManager, signingKeyPrefix string, mesh string) SigningKeyManager {
	return &meshedSigningKeyManager{
		manager:          manager,
		signingKeyPrefix: signingKeyPrefix,
		mesh:             mesh,
	}
}

type meshedSigningKeyManager struct {
	manager          manager.ResourceManager
	signingKeyPrefix string
	mesh             string
}

var _ SigningKeyManager = &meshedSigningKeyManager{}

func (s *meshedSigningKeyManager) GetLatestSigningKey() (*rsa.PrivateKey, int, error) {
	resources := system.SecretResourceList{}
	if err := s.manager.List(context.Background(), &resources, store.ListByMesh(s.mesh)); err != nil {
		return nil, 0, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return latestSigningKey(&resources, s.signingKeyPrefix, s.mesh)
}

func (s *meshedSigningKeyManager) CreateDefaultSigningKey() error {
	return s.CreateSigningKey(DefaultSerialNumber)
}

func (s *meshedSigningKeyManager) CreateSigningKey(serialNumber int) error {
	key, err := NewSigningKey()
	if err != nil {
		return err
	}

	secret := system.NewSecretResource()
	secret.Spec = &system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: key,
		},
	}
	return s.manager.Create(context.Background(), secret, store.CreateBy(SigningKeyResourceKey(s.signingKeyPrefix, serialNumber, s.mesh)))
}
