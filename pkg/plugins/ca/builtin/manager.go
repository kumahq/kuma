package builtin

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	core_ca "github.com/Kong/kuma/pkg/core/ca"
	ca_issuer "github.com/Kong/kuma/pkg/core/ca/issuer"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

type builtinCaManager struct {
	secretManager secret_manager.SecretManager
}

func NewBuiltinCaManager(secretManager secret_manager.SecretManager) core_ca.Manager {
	return &builtinCaManager{
		secretManager: secretManager,
	}
}

var _ core_ca.Manager = &builtinCaManager{}

func (b *builtinCaManager) Ensure(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error {
	_, err := b.getCa(ctx, mesh, backend.Name)
	if core_store.IsResourceNotFound(err) {
		if err := b.create(ctx, mesh, backend.Name); err != nil {
			return errors.Wrapf(err, "failed to create CA for mesh %q and backend %q", mesh, backend.Name)
		}
	} else {
		return err
	}
	return nil
}

func (b *builtinCaManager) ValidateBackend(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error {
	return nil // builtin CA has no config
}

func (b *builtinCaManager) create(ctx context.Context, mesh string, backendName string) error {
	keyPair, err := newRootCa(mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to generate a Root CA cert for Mesh %q", mesh)
	}

	certSecret := &core_system.SecretResource{
		Spec: system_proto.Secret{
			Data: &wrappers.BytesValue{
				Value: keyPair.CertPEM,
			},
		},
	}
	if err := b.secretManager.Create(ctx, certSecret, core_store.CreateBy(certSecretResKey(mesh, backendName))); err != nil {
		return err
	}

	keySecret := &core_system.SecretResource{
		Spec: system_proto.Secret{
			Data: &wrappers.BytesValue{
				Value: keyPair.KeyPEM,
			},
		},
	}
	if err := b.secretManager.Create(ctx, keySecret, core_store.CreateBy(keySecretResKey(mesh, backendName))); err != nil {
		return err
	}
	return nil
}

func certSecretResKey(mesh string, backendName string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Mesh: mesh,
		Name: fmt.Sprintf("%s.ca-builtin-cert-%s", mesh, backendName), // we add mesh as a prefix to have uniqueness of Secret names on K8S
	}
}

func keySecretResKey(mesh string, backendName string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Mesh: mesh,
		Name: fmt.Sprintf("%s.ca-builtin-key-%s", mesh, backendName), // we add mesh as a prefix to have uniqueness of Secret names on K8S
	}
}

func (b *builtinCaManager) GetRootCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) ([]core_ca.Cert, error) {
	ca, err := b.getCa(ctx, mesh, backend.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}
	return []core_ca.Cert{ca.CertPEM}, nil
}

func (b *builtinCaManager) GenerateDataplaneCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend, service string) (core_ca.KeyPair, error) {
	ca, err := b.getCa(ctx, mesh, backend.Name)
	if err != nil {
		return core_ca.KeyPair{}, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}

	var opts []ca_issuer.CertOptsFn
	if backend.GetDpCert().GetRotation().GetExpiration() != nil {
		duration := util_proto.ToDuration(*backend.GetDpCert().GetRotation().Expiration)
		opts = append(opts, ca_issuer.WithExpirationTime(duration))
	}
	keyPair, err := ca_issuer.NewWorkloadCert(ca, mesh, service, opts...)
	if err != nil {
		return core_ca.KeyPair{}, errors.Wrapf(err, "failed to generate a Workload Identity cert for workload %q in Mesh %q using backend %q", service, mesh, backend)
	}
	return *keyPair, nil
}

func (b *builtinCaManager) getCa(ctx context.Context, mesh string, backendName string) (core_ca.KeyPair, error) {
	certSecret := &core_system.SecretResource{}
	if err := b.secretManager.Get(ctx, certSecret, core_store.GetBy(certSecretResKey(mesh, backendName))); err != nil {
		return core_ca.KeyPair{}, err
	}

	keySecret := &core_system.SecretResource{}
	if err := b.secretManager.Get(ctx, keySecret, core_store.GetBy(keySecretResKey(mesh, backendName))); err != nil {
		return core_ca.KeyPair{}, err
	}

	return core_ca.KeyPair{
		CertPEM: certSecret.Spec.Data.Value,
		KeyPEM:  keySecret.Spec.Data.Value,
	}, nil
}
