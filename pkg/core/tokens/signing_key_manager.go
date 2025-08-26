package tokens

import (
	"context"
	"crypto/rsa"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

const (
	DefaultKeyID = "1"
)

// SigningKeyManager manages tokens's signing keys.
// We can have many signing keys in the system.
// Example: "user-token-signing-key-1", "user-token-signing-key-2" etc.
// "user-token-signing-key" has a serial number of 0
// The latest key is  a key with a higher serial number (number at the end of the name)
type SigningKeyManager interface {
	GetLatestSigningKey(context.Context) (*rsa.PrivateKey, string, error)
	CreateDefaultSigningKey(context.Context) error
	CreateSigningKey(ctx context.Context, keyID KeyID) error
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

func (s *signingKeyManager) GetLatestSigningKey(ctx context.Context) (*rsa.PrivateKey, string, error) {
	resources := system.GlobalSecretResourceList{}
	if err := s.manager.List(ctx, &resources); err != nil {
		return nil, "", errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return latestSigningKey(&resources, s.signingKeyPrefix, model.NoMesh)
}

func latestSigningKey(list model.ResourceList, prefix, mesh string) (*rsa.PrivateKey, string, error) {
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
		return nil, "", &SigningKeyNotFound{
			KeyID:  DefaultKeyID,
			Prefix: prefix,
			Mesh:   mesh,
		}
	}

	key, err := keyBytesToRsaPrivateKey(signingKey.GetSpec().(*system_proto.Secret).GetData().GetValue())
	if err != nil {
		return nil, "", err
	}

	return key, strconv.Itoa(highestSerialNumber), nil
}

func (s *signingKeyManager) CreateDefaultSigningKey(ctx context.Context) error {
	return s.CreateSigningKey(ctx, DefaultKeyID)
}

func (s *signingKeyManager) CreateSigningKey(ctx context.Context, keyID KeyID) error {
	key, err := NewSigningKey()
	if err != nil {
		return err
	}

	secret := system.NewGlobalSecretResource()
	secret.Spec = &system_proto.Secret{
		Data: &wrapperspb.BytesValue{
			Value: key,
		},
	}
	return s.manager.Create(ctx, secret, store.CreateBy(SigningKeyResourceKey(s.signingKeyPrefix, keyID, model.NoMesh)))
}
