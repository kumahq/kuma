package issuer

import (
	"context"
	"crypto/rsa"

	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type SigningKeyAccessor interface {
	GetSigningPublicKey(serialNumber int) (*rsa.PublicKey, error)
}

type signingKeyAccessor struct {
	resManager manager.ResourceManager
}

var _ SigningKeyAccessor = &signingKeyAccessor{}

func NewSigningKeyAccessor(resManager manager.ResourceManager) SigningKeyAccessor {
	return &signingKeyAccessor{
		resManager: resManager,
	}
}

func (s *signingKeyAccessor) GetSigningPublicKey(serialNumber int) (*rsa.PublicKey, error) {
	resource := system.NewGlobalSecretResource()
	if err := s.resManager.Get(context.Background(), resource, store.GetBy(SigningKeyResourceKey(serialNumber))); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, SigningKeyNotFound
		}
		return nil, errors.Wrap(err, "could not retrieve signing key")
	}

	key, err := util_rsa.FromPEMBytes(resource.Spec.GetData().GetValue())
	if err != nil {
		return nil, err
	}
	return &key.PublicKey, nil
}
