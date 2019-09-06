package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"

	core_system "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/system"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	secret_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/manager"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/tls"
)

type CaRootCert = []byte

type CaRoot struct {
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}

type BuiltinCa struct {
	Roots []CaRoot `json:"roots"`
}

type BuiltinCaManager interface {
	Create(ctx context.Context, mesh string) error
	Delete(ctx context.Context, mesh string) error
	GetRootCerts(ctx context.Context, mesh string) ([]CaRootCert, error)
}

func NewBuiltinCaManager(secretManager secret_manager.SecretManager) BuiltinCaManager {
	return &builtinCaManager{
		secretManager: secretManager,
	}
}

type builtinCaManager struct {
	secretManager secret_manager.SecretManager
}

func (m *builtinCaManager) Create(ctx context.Context, mesh string) error {
	keyPair, err := tls.NewSelfSignedCert(mesh)
	if err != nil {
		return errors.Wrap(err, "failed to generate a Root CA cert for a given mesh")
	}
	builtinCa := BuiltinCa{
		Roots: []CaRoot{
			{
				Cert: keyPair.CertPEM,
				Key:  keyPair.KeyPEM,
			},
		},
	}
	data, err := json.Marshal(builtinCa)
	if err != nil {
		return errors.Wrap(err, "failed to serialize a Root CA cert for a given mesh")
	}
	builtinCaSecret := &core_system.SecretResource{
		Spec: types.BytesValue{
			Value: data,
		},
	}
	secretKey := builtinCaSecretKey(mesh)
	if err := m.secretManager.Create(ctx, builtinCaSecret, core_store.CreateBy(secretKey)); err != nil {
		return errors.Wrapf(err, "failed to create Builtin CA for a given mesh")
	}
	return nil
}

func (m *builtinCaManager) Delete(ctx context.Context, mesh string) error {
	secretKey := builtinCaSecretKey(mesh)
	builtinCaSecret := &core_system.SecretResource{}
	if err := m.secretManager.Delete(ctx, builtinCaSecret, core_store.DeleteBy(secretKey)); err != nil {
		if core_store.IsResourceNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "failed to delete Builtin CA for a given mesh")
	}
	return nil
}

func (m *builtinCaManager) GetRootCerts(ctx context.Context, mesh string) ([]CaRootCert, error) {
	secretKey := builtinCaSecretKey(mesh)
	builtinCaSecret := &core_system.SecretResource{}
	if err := m.secretManager.Get(ctx, builtinCaSecret, core_store.GetBy(secretKey)); err != nil {
		return nil, err
	}
	builtinCa := BuiltinCa{}
	if err := json.Unmarshal(builtinCaSecret.Spec.Value, &builtinCa); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize a Root CA cert for a given mesh")
	}
	caRootCerts := make([]CaRootCert, len(builtinCa.Roots))
	for i, root := range builtinCa.Roots {
		caRootCerts[i] = root.Cert
	}
	return caRootCerts, nil
}

func builtinCaSecretKey(mesh string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Mesh:      mesh,
		Namespace: core_model.DefaultNamespace,
		Name:      builtinCaSecretName(mesh),
	}
}

func builtinCaSecretName(mesh string) string {
	return fmt.Sprintf("builtinca.%s", mesh)
}
