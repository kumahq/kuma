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
	GetLatestSigningKey(context.Context) (*rsa.PrivateKey, int, error)
	CreateDefaultSigningKey(context.Context) error
	CreateSigningKey(ctx context.Context, serialNumber int) error
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

func (s *signingKeyManager) GetLatestSigningKey(ctx context.Context) (*rsa.PrivateKey, int, error) {
	resources := system.GlobalSecretResourceList{}
	if err := s.manager.List(ctx, &resources); err != nil {
		return nil, 0, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return latestSigningKey(&resources, s.signingKeyPrefix, model.NoMesh)
}

func latestSigningKey(list model.ResourceList, prefix string, mesh string) (*rsa.PrivateKey, int, error) {
	var signingKey model.Resource
	highestSerialNumber := -1
	for _, resource := range list.GetItems() {
		if !strings.HasPrefix(resource.GetMeta().GetName(), prefix) {
			continue
		}
		serialNumber, _ := signingKeySerialNumber(resource.GetMeta().GetName(), prefix)
		if serialNumber > highestSerialNumber {
			signingKey = resource
			highestSerialNumber = serialNumber
		}
	}

	if signingKey == nil {
		return nil, 0, &SigningKeyNotFound{
			SerialNumber: DefaultSerialNumber,
			Prefix:       prefix,
			Mesh:         mesh,
		}
	}

	key, err := keyBytesToRsaPrivateKey(signingKey.GetSpec().(*system_proto.Secret).GetData().GetValue())
	if err != nil {
		return nil, 0, err
	}

	return key, highestSerialNumber, nil
}

func (s *signingKeyManager) CreateDefaultSigningKey(ctx context.Context) error {
	return s.CreateSigningKey(ctx, DefaultSerialNumber)
}

func (s *signingKeyManager) CreateSigningKey(ctx context.Context, serialNumber int) error {
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
	return s.manager.Create(ctx, secret, store.CreateBy(SigningKeyResourceKey(s.signingKeyPrefix, serialNumber, model.NoMesh)))
}
