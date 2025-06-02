package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
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

func (s *meshedSigningKeyManager) GetLatestSigningKey(ctx context.Context) (*rsa.PrivateKey, KeyID, error) {
	resources := system.SecretResourceList{}
	if err := s.manager.List(ctx, &resources, store.ListByMesh(s.mesh)); err != nil {
		return nil, "", errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return latestSigningKey(&resources, s.signingKeyPrefix, s.mesh)
}

func (s *meshedSigningKeyManager) CreateDefaultSigningKey(ctx context.Context) error {
	return s.CreateSigningKey(ctx, DefaultKeyID)
}

func (s *meshedSigningKeyManager) CreateSigningKey(ctx context.Context, keyID KeyID) error {
	key, err := NewSigningKey()
	if err != nil {
		return err
	}

	secret := system.NewSecretResource()
	secret.Spec = &system_proto.Secret{
		Data: &wrapperspb.BytesValue{
			Value: key,
		},
	}

	owner := core_mesh.NewMeshResource()
	if err := s.manager.Get(ctx, owner, store.GetByKey(s.mesh, model.NoMesh), store.GetConsistent()); err != nil {
		return manager.MeshNotFound(s.mesh)
	}

	return s.manager.Create(ctx, secret, store.CreateWithOwner(owner), store.CreateBy(SigningKeyResourceKey(s.signingKeyPrefix, keyID, s.mesh)))
}
