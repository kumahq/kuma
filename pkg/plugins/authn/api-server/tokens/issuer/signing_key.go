package issuer

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

const (
	signingKeyPrefix = "user-token-signing-key-"

	DefaultSerialNumber = 1
)

var SigningKeyNotFound = errors.New("there is no signing key in the Control Plane")

func SigningKeyResourceKey(serialNumber int) model.ResourceKey {
	return model.ResourceKey{
		Name: fmt.Sprintf("%s%d", signingKeyPrefix, serialNumber),
	}
}

func IsSigningKeyResource(resKey model.ResourceKey) bool {
	return strings.HasPrefix(resKey.Name, signingKeyPrefix) && resKey.Mesh == ""
}

// SigningKeyManager manages User Token's signing keys.
// We can have many signing keys in the system.
// Example: "user-token-signing-key-1", "user-token-signing-key-2" etc.
// The latest key is  a key with a higher serial number (number at the end of the name)
type SigningKeyManager interface {
	GetSigningKey(serialNumber int) ([]byte, error)
	GetLatestSigningKey() ([]byte, int, error)
	CreateDefaultSigningKey() error
	CreateSigningKey(serialNumber int) error
}

func NewSigningKeyManager(manager manager.ResourceManager) SigningKeyManager {
	return &signingKeyManager{
		manager: manager,
	}
}

type signingKeyManager struct {
	manager manager.ResourceManager
}

var _ SigningKeyManager = &signingKeyManager{}

func (s *signingKeyManager) GetSigningKey(serialNumber int) ([]byte, error) {
	resource := system.NewGlobalSecretResource()
	if err := s.manager.Get(context.Background(), resource, store.GetBy(SigningKeyResourceKey(serialNumber))); err != nil {
		if store.IsResourceNotFound(err) {
			return nil, SigningKeyNotFound
		}
		return nil, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}
	return resource.Spec.GetData().GetValue(), nil
}

func (s *signingKeyManager) GetLatestSigningKey() ([]byte, int, error) {
	resources := system.GlobalSecretResourceList{}
	if err := s.manager.List(context.Background(), &resources); err != nil {
		return nil, 0, errors.Wrap(err, "could not retrieve signing key from secret manager")
	}

	var signingKeys []*system.GlobalSecretResource
	for _, resource := range resources.Items {
		if strings.HasPrefix(resource.Meta.GetName(), signingKeyPrefix) {
			signingKeys = append(signingKeys, resource)
		}
	}

	if len(signingKeys) == 0 {
		return nil, 0, SigningKeyNotFound
	}

	sort.Stable(GlobalSecretsBySerial(signingKeys))

	serialNumber, err := signingKeySerialNumber(signingKeys[0].Meta.GetName())
	if err != nil {
		return nil, 0, err
	}

	return signingKeys[0].Spec.GetData().GetValue(), serialNumber, nil
}

func (s *signingKeyManager) CreateDefaultSigningKey() error {
	return s.CreateSigningKey(DefaultSerialNumber)
}

func (s *signingKeyManager) CreateSigningKey(serialNumber int) error {
	key, err := issuer.NewSigningKey()
	if err != nil {
		return errors.Wrap(err, "could not construct signing key")
	}

	secret := system.NewGlobalSecretResource()
	secret.Spec = &system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: key,
		},
	}
	if err := s.manager.Create(context.Background(), secret, store.CreateBy(SigningKeyResourceKey(serialNumber))); err != nil {
		return errors.Wrap(err, "could not create signing key")
	}
	return nil
}

func signingKeySerialNumber(secretName string) (int, error) {
	serialNumberStr := strings.ReplaceAll(secretName, signingKeyPrefix, "")
	serialNumber, err := strconv.Atoi(serialNumberStr)
	if err != nil {
		return 0, err
	}
	return serialNumber, nil
}

type GlobalSecretsBySerial []*system.GlobalSecretResource

func (a GlobalSecretsBySerial) Len() int      { return len(a) }
func (a GlobalSecretsBySerial) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a GlobalSecretsBySerial) Less(i, j int) bool {
	// ignore errors and assume serial number is 0 when secret has wrong format
	iSerialNumber, _ := signingKeySerialNumber(a[i].Meta.GetName())
	jSerialNumber, _ := signingKeySerialNumber(a[j].Meta.GetName())
	return iSerialNumber > jSerialNumber
}
