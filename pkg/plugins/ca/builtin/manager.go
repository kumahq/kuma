package builtin

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca"
	ca_issuer "github.com/Kong/kuma/pkg/core/ca/issuer"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
)

type builtinCaManager struct {
	secretManager secret_manager.SecretManager
}

func NewBuiltinCaManager(secretManager secret_manager.SecretManager) ca.CaManager {
	return &builtinCaManager{
		secretManager: secretManager,
	}
}

var _ ca.CaManager = &builtinCaManager{}

const certSecretPattern = "ca-builtin-cert-%s"
const keySecretPattern = "ca-builtin-key-%s"

func (b *builtinCaManager) Ensure(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error {
	_, err := b.getCa(ctx, mesh, backend.Name)
	if core_store.IsResourceNotFound(err) {
		err = b.create(ctx, mesh, backend.Name)
	}
	return err
}

func (b *builtinCaManager) ValidateBackend(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error {
	return nil
}

func (b *builtinCaManager) create(ctx context.Context, mesh string, backendName string) error {
	keyPair, err := ca_issuer.NewRootCA(mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to generate a Root CA cert for Mesh %q", mesh)
	}

	certSecretName := fmt.Sprintf(certSecretPattern, backendName)
	certSecret := &core_system.SecretResource{
		Spec: system_proto.Secret{
			Data: &wrappers.BytesValue{
				Value: keyPair.CertPEM,
			},
		},
	}
	if err := b.secretManager.Create(ctx, certSecret, core_store.CreateByKey(certSecretName, mesh)); err != nil {
		return err
	}

	keySecretName := fmt.Sprintf(keySecretPattern, backendName)
	keySecret := &core_system.SecretResource{
		Spec: system_proto.Secret{
			Data: &wrappers.BytesValue{
				Value: keyPair.KeyPEM,
			},
		},
	}
	if err := b.secretManager.Create(ctx, keySecret, core_store.CreateByKey(keySecretName, mesh)); err != nil {
		return err
	}
	return nil

}

func (b *builtinCaManager) GetRootCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) (ca.Cert, error) {
	ca, err := b.getCa(ctx, mesh, backend.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}
	return ca.CertPEM, nil
}

func (b *builtinCaManager) GenerateDataplaneCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend, service string) (ca.KeyPair, error) {
	meshCa, err := b.getCa(ctx, mesh, backend.Name)
	if err != nil {
		return ca.KeyPair{}, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}

	keyPair, err := ca_issuer.NewWorkloadCert(meshCa, mesh, service)
	if err != nil {
		return ca.KeyPair{}, errors.Wrapf(err, "failed to generate a Workload Identity cert for workload %q in Mesh %q using backend %q", service, mesh, backend)
	}
	return *keyPair, nil // todo pointer?
}

func (b *builtinCaManager) getCa(ctx context.Context, mesh string, backendName string) (ca.KeyPair, error) {
	certSecretName := fmt.Sprintf(certSecretPattern, backendName)
	certSecret := &core_system.SecretResource{}
	if err := b.secretManager.Get(ctx, certSecret, core_store.GetByKey(certSecretName, mesh)); err != nil {
		return ca.KeyPair{}, err
	}

	keySecretName := fmt.Sprintf(keySecretPattern, backendName)
	keySecret := &core_system.SecretResource{}
	if err := b.secretManager.Get(ctx, keySecret, core_store.GetByKey(keySecretName, mesh)); err != nil {
		return ca.KeyPair{}, err
	}

	return ca.KeyPair{
		CertPEM: certSecret.Spec.Data.Value,
		KeyPEM:  keySecret.Spec.Data.Value,
	}, nil
}
