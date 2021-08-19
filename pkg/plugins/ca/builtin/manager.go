package builtin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	ca_issuer "github.com/kumahq/kuma/pkg/core/ca/issuer"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_validators "github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/ca/builtin/config"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type builtinCaManager struct {
	secretManager manager.ResourceManager
}

func NewBuiltinCaManager(secretManager manager.ResourceManager) core_ca.Manager {
	return &builtinCaManager{
		secretManager: secretManager,
	}
}

var _ core_ca.Manager = &builtinCaManager{}

func (b *builtinCaManager) Ensure(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) error {
	_, err := b.getCa(ctx, mesh, backend.Name)
	if core_store.IsResourceNotFound(err) {
		if err := b.create(ctx, mesh, backend); err != nil {
			return errors.Wrapf(err, "failed to create CA for mesh %q and backend %q", mesh, backend.Name)
		}
	} else {
		return err
	}
	return nil
}

func (b *builtinCaManager) ValidateBackend(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) error {
	verr := core_validators.ValidationError{}
	cfg := &config.BuiltinCertificateAuthorityConfig{}
	if err := util_proto.ToTyped(backend.Conf, cfg); err != nil {
		verr.AddViolation("", "could not convert backend config: "+err.Error())
		return verr.OrNil()
	}
	return nil
}

func (b *builtinCaManager) UsedSecrets(mesh string, backend *mesh_proto.CertificateAuthorityBackend) ([]string, error) {
	return []string{
		certSecretResKey(mesh, backend.Name).Name,
		keySecretResKey(mesh, backend.Name).Name,
	}, nil
}

func (b *builtinCaManager) create(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) error {
	cfg := &config.BuiltinCertificateAuthorityConfig{}
	if err := util_proto.ToTyped(backend.Conf, cfg); err != nil {
		return errors.Wrap(err, "could not convert backend config to BuiltinCertificateAuthorityConfig")
	}

	var opts []certOptsFn
	if cfg.GetCaCert().GetExpiration() != "" {
		duration, err := core_mesh.ParseDuration(cfg.GetCaCert().GetExpiration())
		if err != nil {
			return err
		}
		opts = append(opts, withExpirationTime(duration))
	}
	keyPair, err := newRootCa(mesh, int(cfg.GetCaCert().GetRSAbits().GetValue()), opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to generate a Root CA cert for Mesh %q", mesh)
	}

	certSecret := &core_system.SecretResource{
		Spec: &system_proto.Secret{
			Data: util_proto.Bytes(keyPair.CertPEM),
		},
	}
	if err := b.secretManager.Create(ctx, certSecret, core_store.CreateBy(certSecretResKey(mesh, backend.Name))); err != nil {
		return err
	}

	keySecret := &core_system.SecretResource{
		Spec: &system_proto.Secret{
			Data: util_proto.Bytes(keyPair.KeyPEM),
		},
	}
	if err := b.secretManager.Create(ctx, keySecret, core_store.CreateBy(keySecretResKey(mesh, backend.Name))); err != nil {
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

func (b *builtinCaManager) GetRootCert(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) ([]core_ca.Cert, error) {
	ca, err := b.getCa(ctx, mesh, backend.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}
	return []core_ca.Cert{ca.CertPEM}, nil
}

func (b *builtinCaManager) GenerateDataplaneCert(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend, tags mesh_proto.MultiValueTagSet) (core_ca.KeyPair, error) {
	ca, err := b.getCa(ctx, mesh, backend.Name)
	if err != nil {
		return core_ca.KeyPair{}, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}

	var opts []ca_issuer.CertOptsFn
	if backend.GetDpCert().GetRotation().GetExpiration() != "" {
		duration, err := core_mesh.ParseDuration(backend.GetDpCert().GetRotation().Expiration)
		if err != nil {
			return core_ca.KeyPair{}, err
		}
		opts = append(opts, ca_issuer.WithExpirationTime(duration))
	}
	keyPair, err := ca_issuer.NewWorkloadCert(ca, mesh, tags, opts...)
	if err != nil {
		return core_ca.KeyPair{}, errors.Wrapf(err, "failed to generate a Workload Identity cert for tags %q in Mesh %q using backend %q", tags.String(), mesh, backend)
	}
	return *keyPair, nil
}

func (b *builtinCaManager) getCa(ctx context.Context, mesh string, backendName string) (core_ca.KeyPair, error) {
	certSecret := core_system.NewSecretResource()
	if err := b.secretManager.Get(ctx, certSecret, core_store.GetBy(certSecretResKey(mesh, backendName))); err != nil {
		return core_ca.KeyPair{}, err
	}

	keySecret := core_system.NewSecretResource()
	if err := b.secretManager.Get(ctx, keySecret, core_store.GetBy(keySecretResKey(mesh, backendName))); err != nil {
		return core_ca.KeyPair{}, err
	}

	return core_ca.KeyPair{
		CertPEM: certSecret.Spec.Data.Value,
		KeyPEM:  keySecret.Spec.Data.Value,
	}, nil
}
