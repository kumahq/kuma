package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
)

// signingKeyFromPublicKeyAccessor is an accessor which can be used in situation
// where the secret contains only public key part of signing key, like
// in zone token, where only global contains private key, and zones only
// public ones to validate tokens
type signingKeyFromPublicKeyAccessor struct {
	resManager       manager.ReadOnlyResourceManager
	signingKeyPrefix string
}

var _ SigningKeyAccessor = &signingKeyFromPublicKeyAccessor{}

func NewSigningKeyFromPublicKeyAccessor(resManager manager.ReadOnlyResourceManager, signingKeyPrefix string) SigningKeyAccessor {
	return &signingKeyFromPublicKeyAccessor{
		resManager:       resManager,
		signingKeyPrefix: signingKeyPrefix,
	}
}

func (s *signingKeyFromPublicKeyAccessor) GetPublicKey(ctx context.Context, keyID KeyID) (*rsa.PublicKey, error) {
	keyBytes, err := s.getKeyBytes(ctx, keyID)
	if err != nil {
		return nil, err
	}

	return keyBytesToRsaPublicKey(keyBytes)
}

func (s *signingKeyFromPublicKeyAccessor) getKeyBytes(ctx context.Context, keyID KeyID) ([]byte, error) {
	return getKeyBytes(ctx, s.resManager, s.signingKeyPrefix, keyID)
}

// GetLegacyKey is not supported for this accessor as it's not used for signing
// keys from pre 1.4.x version of Kuma, where we used symmetric HMAC256 method of signing DP keys.
func (s *signingKeyFromPublicKeyAccessor) GetLegacyKey(_ context.Context, _ string) ([]byte, error) {
	return nil, errors.New("legacy key are not supported")
}
