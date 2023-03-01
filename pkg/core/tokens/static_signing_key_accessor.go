package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/pkg/errors"

	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

type staticSigningKeyAccessor struct {
	keys map[string]*rsa.PublicKey
}

var _ SigningKeyAccessor = &staticSigningKeyAccessor{}

func NewStaticSigningKeyAccessor(keys []PublicKey) (SigningKeyAccessor, error) {
	s := &staticSigningKeyAccessor{
		keys: map[string]*rsa.PublicKey{},
	}
	for _, key := range keys {
		publicKey, err := util_rsa.FromPEMBytesToPublicKey([]byte(key.PEM))
		if err != nil {
			return nil, err
		}
		s.keys[key.KID] = publicKey
	}
	return s, nil
}

func (s *staticSigningKeyAccessor) GetPublicKey(ctx context.Context, keyID KeyID) (*rsa.PublicKey, error) {
	key, ok := s.keys[keyID]
	if !ok {
		return nil, &SigningKeyNotFound{
			KeyID: keyID,
		}
	}
	return key, nil
}

func (s *staticSigningKeyAccessor) GetLegacyKey(context.Context, KeyID) ([]byte, error) {
	return nil, errors.New("not supported")
}
