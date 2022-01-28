package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

// SigningKeyAccessor access public part of signing key
// In the future, we may add offline token generation (kumactl without CP running or external system)
// In that case, we could provide only public key to the CP via static configuration.
// So we can easily do this by providing separate implementation for this interface.
type SigningKeyAccessor interface {
	GetPublicKey(ctx context.Context, serialNumber int) (*rsa.PublicKey, error)
	// GetLegacyKey returns legacy key. In pre 1.4.x version of Kuma, we used symmetric HMAC256 method of signing DP keys.
	// In that case, we have to retrieve private key even for verification.
	GetLegacyKey(ctx context.Context, serialNumber int) ([]byte, error)
}

type signingKeyAccessor struct {
	resManager       manager.ResourceManager
	signingKeyPrefix string
	isPublic         bool
}

var _ SigningKeyAccessor = &signingKeyAccessor{}

func NewSigningKeyAccessor(resManager manager.ResourceManager, signingKeyPrefix string, isPublic bool) SigningKeyAccessor {
	return &signingKeyAccessor{
		resManager:       resManager,
		signingKeyPrefix: signingKeyPrefix,
		isPublic:         isPublic,
	}
}

func (s *signingKeyAccessor) GetPublicKey(ctx context.Context, serialNumber int) (*rsa.PublicKey, error) {
	keyBytes, err := s.getKeyBytes(ctx, serialNumber)
	if err != nil {
		return nil, err
	}

	if s.isPublic {
		return keyBytesToRsaPublicKey(keyBytes)
	}

	key, err := keyBytesToRsaPrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return &key.PublicKey, nil
}

func (s *signingKeyAccessor) getKeyBytes(ctx context.Context, serialNumber int) ([]byte, error) {
	resource := system.NewGlobalSecretResource()
	if err := s.resManager.Get(ctx, resource, store.GetBy(SigningKeyResourceKey(s.signingKeyPrefix, serialNumber, model.NoMesh))); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, &SigningKeyNotFound{
				SerialNumber: serialNumber,
				Prefix:       s.signingKeyPrefix,
			}
		}
		return nil, errors.Wrap(err, "could not retrieve signing key")
	}
	return resource.Spec.GetData().GetValue(), nil
}

func (s *signingKeyAccessor) GetLegacyKey(ctx context.Context, serialNumber int) ([]byte, error) {
	return s.getKeyBytes(ctx, serialNumber)
}
