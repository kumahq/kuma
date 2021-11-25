package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type meshedSigningKeyAccessor struct {
	resManager       manager.ResourceManager
	signingKeyPrefix string
	mesh             string
}

var _ SigningKeyAccessor = &meshedSigningKeyAccessor{}

// NewMeshedSigningKeyAccessor builds SigningKeyAccessor that is bound to a Mesh.
// Some tokens like Dataplane Token are bound to a mesh.
// In this case, singing key is also stored as a Secret in the Mesh, not as GlobalSecret.
func NewMeshedSigningKeyAccessor(resManager manager.ResourceManager, signingKeyPrefix string, mesh string) SigningKeyAccessor {
	return &meshedSigningKeyAccessor{
		resManager:       resManager,
		signingKeyPrefix: signingKeyPrefix,
		mesh:             mesh,
	}
}

func (s *meshedSigningKeyAccessor) GetPublicKey(serialNumber int) (*rsa.PublicKey, error) {
	keyBytes, err := s.getKeyBytes(serialNumber)
	if err != nil {
		return nil, err
	}
	key, err := keyBytesToRsaKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return &key.PublicKey, nil
}

func (s *meshedSigningKeyAccessor) getKeyBytes(serialNumber int) ([]byte, error) {
	resource := system.NewSecretResource()
	if err := s.resManager.Get(context.Background(), resource, store.GetBy(SigningKeyResourceKey(s.signingKeyPrefix, serialNumber, s.mesh))); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, &SigningKeyNotFound{
				SerialNumber: serialNumber,
				Prefix:       s.signingKeyPrefix,
				Mesh:         s.mesh,
			}
		}
		return nil, errors.Wrap(err, "could not retrieve signing key")
	}
	return resource.Spec.GetData().GetValue(), nil
}

func (s *meshedSigningKeyAccessor) GetLegacyKey(serialNumber int) ([]byte, error) {
	return s.getKeyBytes(serialNumber)
}
