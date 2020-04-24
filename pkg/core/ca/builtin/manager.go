package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	builtin_issuer "github.com/Kong/kuma/pkg/core/ca/builtin/issuer"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/Kong/kuma/pkg/tls"
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
	Ensure(ctx context.Context, mesh string) error
	Create(ctx context.Context, mesh string) error
	Delete(ctx context.Context, mesh string) error
	GetRootCerts(ctx context.Context, mesh string) ([]CaRootCert, error)
	GenerateWorkloadCert(ctx context.Context, mesh string, workload string) (*tls.KeyPair, error)

	GetSecretName(mesh string) string
}

func NewBuiltinCaManager(secretManager secret_manager.SecretManager) BuiltinCaManager {
	return &builtinCaManager{
		secretManager: secretManager,
	}
}

type builtinCaManager struct {
	secretManager secret_manager.SecretManager
}

func (m *builtinCaManager) Ensure(ctx context.Context, mesh string) error {
	_, err := m.getMeshCa(ctx, mesh)
	if core_store.IsResourceNotFound(err) {
		err = m.Create(ctx, mesh)
	}
	return err
}

func (m *builtinCaManager) Create(ctx context.Context, mesh string) error {
	keyPair, err := builtin_issuer.NewRootCA(mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to generate a Root CA cert for Mesh %q", mesh)
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
		return errors.Wrapf(err, "failed to serialize a Root CA cert for Mesh %q", mesh)
	}
	builtinCaSecret := &core_system.SecretResource{
		Spec: system_proto.Secret{
			Data: &wrappers.BytesValue{
				Value: data,
			},
		},
	}
	secretKey := builtinCaSecretKey(mesh)
	if err := m.secretManager.Create(ctx, builtinCaSecret, core_store.CreateBy(secretKey)); err != nil {
		return errors.Wrapf(err, "failed to create Builtin CA for Mesh %q", mesh)
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
		return errors.Wrapf(err, "failed to delete Builtin CA for Mesh %q", mesh)
	}
	return nil
}

func (m *builtinCaManager) GetRootCerts(ctx context.Context, mesh string) ([]CaRootCert, error) {
	meshCa, err := m.getMeshCa(ctx, mesh)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q", mesh)
	}
	caRootCerts := make([]CaRootCert, len(meshCa.Roots))
	for i, root := range meshCa.Roots {
		caRootCerts[i] = root.Cert
	}
	return caRootCerts, nil
}

func (m *builtinCaManager) GenerateWorkloadCert(ctx context.Context, mesh string, workload string) (*tls.KeyPair, error) {
	meshCa, err := m.getMeshCa(ctx, mesh)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q", mesh)
	}
	if len(meshCa.Roots) < 1 {
		return nil, errors.Wrapf(err, "CA for Mesh %q has no key pair", mesh)
	}
	active := meshCa.Roots[0]
	signer := tls.KeyPair{CertPEM: active.Cert, KeyPEM: active.Key}
	keyPair, err := builtin_issuer.NewWorkloadCert(signer, mesh, workload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate a Workload Identity cert for workload %q in Mesh %q", workload, mesh)
	}
	return keyPair, nil
}

func (m *builtinCaManager) getMeshCa(ctx context.Context, mesh string) (*BuiltinCa, error) {
	secretKey := builtinCaSecretKey(mesh)
	builtinCaSecret := &core_system.SecretResource{}
	if err := m.secretManager.Get(ctx, builtinCaSecret, core_store.GetBy(secretKey)); err != nil {
		return nil, err
	}
	builtinCa := BuiltinCa{}
	if err := json.Unmarshal(builtinCaSecret.Spec.Data.Value, &builtinCa); err != nil {
		return nil, errors.Wrapf(err, "failed to deserialize a Root CA cert for Mesh %q", mesh)
	}
	return &builtinCa, nil
}

func (m *builtinCaManager) GetSecretName(mesh string) string {
	return builtinCaSecretKey(mesh).Name
}

func builtinCaSecretKey(mesh string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Mesh: mesh,
		Name: builtinCaSecretName(mesh),
	}
}

func builtinCaSecretName(mesh string) string {
	return fmt.Sprintf("builtinca.%s", mesh)
}
