package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
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
}

var _ SigningKeyAccessor = &signingKeyAccessor{}

func NewSigningKeyAccessor(resManager manager.ResourceManager, signingKeyPrefix string) SigningKeyAccessor {
	return &signingKeyAccessor{
		resManager:       resManager,
		signingKeyPrefix: signingKeyPrefix,
	}
}

func (s *signingKeyAccessor) GetPublicKey(ctx context.Context, serialNumber int) (*rsa.PublicKey, error) {
	keyBytes, err := getKeyBytes(ctx, s.resManager, s.signingKeyPrefix, serialNumber)
	if err != nil {
		return nil, err
	}

	key, err := keyBytesToRsaPrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return &key.PublicKey, nil
}

func (s *signingKeyAccessor) GetLegacyKey(ctx context.Context, serialNumber int) ([]byte, error) {
	return getKeyBytes(ctx, s.resManager, s.signingKeyPrefix, serialNumber)
}
