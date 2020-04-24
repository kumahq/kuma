package provided

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/ptypes/wrappers"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	builtin_issuer "github.com/Kong/kuma/pkg/core/ca/builtin/issuer"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/Kong/kuma/pkg/tls"
)

type SigningCert struct {
	Id   string `json:"id"`
	Cert []byte `json:"cert"`
}

type SigningKeyCert struct {
	SigningCert
	Key []byte `json:"key"`
}

type ProvidedCa struct {
	SigningKeyCerts []SigningKeyCert `json:"signingKeyCerts"`
}

type ProvidedCaManager interface {
	AddSigningCert(ctx context.Context, mesh string, root tls.KeyPair) (*SigningCert, error)
	DeleteSigningCert(ctx context.Context, mesh string, id string) error

	DeleteCa(ctx context.Context, mesh string) error

	GetSigningCerts(ctx context.Context, mesh string) ([]SigningCert, error)
	GenerateWorkloadCert(ctx context.Context, mesh string, workload string) (*tls.KeyPair, error)
}

type providedCaManager struct {
	secretManager secret_manager.SecretManager
}

var _ ProvidedCaManager = &providedCaManager{}

func NewProvidedCaManager(secretManager secret_manager.SecretManager) ProvidedCaManager {
	return &providedCaManager{secretManager}
}

func (p *providedCaManager) AddSigningCert(ctx context.Context, mesh string, signingPair tls.KeyPair) (*SigningCert, error) {
	if err := ValidateCaCert(signingPair); err != nil {
		return nil, err
	}

	providedCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Get(ctx, providedCaSecret, core_store.GetBy(providedCaSecretKey(mesh))); err != nil {
		if core_store.IsResourceNotFound(err) {
			if err := p.secretManager.Create(ctx, providedCaSecret, core_store.CreateBy(providedCaSecretKey(mesh))); err != nil {
				return nil, errors.Wrapf(err, "could not create provided CA for Mesh %q", mesh)
			}
		} else {
			return nil, errors.Wrapf(err, "failed to load provided CA for Mesh %q", mesh)
		}
	}

	providedCa := ProvidedCa{}
	if len(providedCaSecret.Spec.GetData().GetValue()) > 0 {
		if err := json.Unmarshal(providedCaSecret.Spec.GetData().GetValue(), &providedCa); err != nil {
			return nil, errors.Wrapf(err, "failed to deserialize provided CA for Mesh %q", mesh)
		}
	}

	if len(providedCa.SigningKeyCerts) > 0 {
		return nil, errors.New("cannot add more than 1 signing cert to provided CA")
	}

	signingCert := SigningCert{
		Id:   core.NewUUID(),
		Cert: signingPair.CertPEM,
	}
	signingKeyCert := SigningKeyCert{
		Key:         signingPair.KeyPEM,
		SigningCert: signingCert,
	}
	providedCa.SigningKeyCerts = append(providedCa.SigningKeyCerts, signingKeyCert)

	caBytes, err := json.Marshal(providedCa)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal provided CA")
	}

	providedCaSecret.Spec = system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: caBytes,
		},
	}
	if err := p.secretManager.Update(ctx, providedCaSecret); err != nil {
		return nil, errors.Wrapf(err, "failed to update provided CA for Mesh %q", mesh)
	}
	return &signingCert, nil
}

func (p *providedCaManager) DeleteSigningCert(ctx context.Context, mesh string, id string) error {
	providedCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Get(ctx, providedCaSecret, core_store.GetBy(providedCaSecretKey(mesh))); err != nil {
		return errors.Wrapf(err, "failed to load provided CA for Mesh %q", mesh)
	}
	providedCa := ProvidedCa{}
	if err := json.Unmarshal(providedCaSecret.Spec.GetData().GetValue(), &providedCa); err != nil {
		return errors.Wrapf(err, "failed to deserialize provided CA for Mesh %q", mesh)
	}

	var retainedCaRoots []SigningKeyCert
	for _, root := range providedCa.SigningKeyCerts {
		if root.Id != id {
			retainedCaRoots = append(retainedCaRoots, root)
		}
	}

	if len(retainedCaRoots) == len(providedCa.SigningKeyCerts) {
		return &SigningCertNotFound{
			Id:   id,
			Mesh: mesh,
		}
	}

	providedCa.SigningKeyCerts = retainedCaRoots
	newBytes, err := json.Marshal(providedCa)
	if err != nil {
		return err
	}

	providedCaSecret.Spec = system_proto.Secret{
		Data: &wrappers.BytesValue{
			Value: newBytes,
		},
	}
	if err := p.secretManager.Update(ctx, providedCaSecret); err != nil {
		return errors.Wrapf(err, "failed to update provided CA for Mesh %q", mesh)
	}
	return nil
}

type SigningCertNotFound struct {
	Id   string
	Mesh string
}

func (s *SigningCertNotFound) Error() string {
	return fmt.Sprintf("could not find signing cert with id %q for Mesh %q", s.Id, s.Mesh)
}

func (p *providedCaManager) DeleteCa(ctx context.Context, mesh string) error {
	// If we ever expose this via API, we need to validate that mesh is disabled or the type is other than provided
	secretKey := providedCaSecretKey(mesh)
	caSecret := &core_system.SecretResource{}
	if err := p.secretManager.Delete(ctx, caSecret, core_store.DeleteBy(secretKey)); err != nil {
		if core_store.IsResourceNotFound(err) {
			return err
		}
		return errors.Wrapf(err, "failed to delete provided CA for Mesh %q", mesh)
	}
	return nil
}

func (p *providedCaManager) GetSigningCerts(ctx context.Context, mesh string) ([]SigningCert, error) {
	meshCa, err := p.getMeshCa(ctx, mesh)
	if err != nil {
		if core_store.IsResourceNotFound(err) {
			return nil, err
		}
		return nil, errors.Wrapf(err, "failed to load provided CA for Mesh %q", mesh)
	}
	caRootCerts := make([]SigningCert, len(meshCa.SigningKeyCerts))
	for i, root := range meshCa.SigningKeyCerts {
		caRootCerts[i] = SigningCert{
			Id:   root.Id,
			Cert: root.Cert,
		}
	}
	return caRootCerts, nil
}

func (p *providedCaManager) GenerateWorkloadCert(ctx context.Context, mesh string, workload string) (*tls.KeyPair, error) {
	meshCa, err := p.getMeshCa(ctx, mesh)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load provided CA for Mesh %q", mesh)
	}
	if len(meshCa.SigningKeyCerts) < 1 {
		return nil, errors.Wrapf(err, "provided CA for Mesh %q has no key pair", mesh)
	}
	active := meshCa.SigningKeyCerts[0]
	signer := tls.KeyPair{CertPEM: active.Cert, KeyPEM: active.Key}
	keyPair, err := builtin_issuer.NewWorkloadCert(signer, mesh, workload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate a Workload Identity cert for workload %q in Mesh %q", workload, mesh)
	}
	return keyPair, nil
}

func (p *providedCaManager) getMeshCa(ctx context.Context, mesh string) (*ProvidedCa, error) {
	secretKey := providedCaSecretKey(mesh)
	providedCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Get(ctx, providedCaSecret, core_store.GetBy(secretKey)); err != nil {
		return nil, err
	}
	providedCa := ProvidedCa{}
	if err := json.Unmarshal(providedCaSecret.Spec.GetData().GetValue(), &providedCa); err != nil {
		return nil, errors.Wrapf(err, "failed to deserialize provided CA for Mesh %q", mesh)
	}
	return &providedCa, nil
}

func providedCaSecretKey(mesh string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Mesh: mesh,
		Name: providedCaSecretName(mesh),
	}
}

func providedCaSecretName(mesh string) string {
	return fmt.Sprintf("providedca.%s", mesh)
}
