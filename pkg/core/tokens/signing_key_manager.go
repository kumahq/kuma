package tokens

import (
	"context"
	"crypto/rsa"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

const (
	DefaultSerialNumber = 1
)

// SigningKeyManager manages tokens's signing keys.
// We can have many signing keys in the system.
// Example: "user-token-signing-key-1", "user-token-signing-key-2" etc.
// "user-token-signing-key" has a serial number of 0
// The latest key is  a key with a higher serial number (number at the end of the name)
type SigningKeyManager interface {
	GetLatestSigningKey() (*rsa.PrivateKey, int, error)
	CreateDefaultSigningKey() error
	CreateSigningKey(serialNumber int) error
}

func NewSigningKeyManager(manager manager.ResourceManager, signingKeyPrefix string) SigningKeyManager {
	return &signingKeyManager{
		manager:          manager,
		signingKeyPrefix: signingKeyPrefix,
	}
}

type signingKeyManager struct {
	manager          manager.ResourceManager
	signingKeyPrefix string
}

var _ SigningKeyManager = &signingKeyManager{}

func (s *signingKeyManager) GetLatestSigningKey() (*rsa.PrivateKey, int, error) {
	resources := system.GlobalSecretResourceList{}
	if err := s.manager.List(context.Background(), &resources); err != nil {
		return nil, 0, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}

	var signingKey *system.GlobalSecretResource
	highestSerialNumber := -1
	for _, resource := range resources.Items {
		if !strings.HasPrefix(resource.Meta.GetName(), s.signingKeyPrefix) {
			continue
		}
		serialNumber, _ := signingKeySerialNumber(resource.Meta.GetName(), s.signingKeyPrefix)
		if serialNumber > highestSerialNumber {
			signingKey = resource
			highestSerialNumber = serialNumber
		}
	}

	if signingKey == nil {
		return nil, 0, &SigningKeyNotFound{SerialNumber: DefaultSerialNumber, Prefix: s.signingKeyPrefix}
	}

	key, err := keyBytesToRsaKey(signingKey.Spec.GetData().GetValue())
	if err != nil {
		return nil, 0, err
	}
	return key, highestSerialNumber, nil
}

func (s *signingKeyManager) CreateDefaultSigningKey() error {
	return s.CreateSigningKey(DefaultSerialNumber)
}

func (s *signingKeyManager) CreateSigningKey(serialNumber int) error {
	key, err := NewSigningKey()
	if err != nil {
		return err
	}

	secret := system.NewGlobalSecretResource()
	secret.Spec = &system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: key,
		},
	}
	return s.manager.Create(context.Background(), secret, store.CreateBy(SigningKeyResourceKey(s.signingKeyPrefix, serialNumber, model.NoMesh)))
}
