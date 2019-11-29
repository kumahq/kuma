package provided

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/kuma/pkg/tls"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	builtin_issuer "github.com/Kong/kuma/pkg/core/ca/builtin/issuer"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
)

type CaRootCert struct {
	Id   string `json:"id"`
	Cert []byte `json:"cert"`
}

type CaRoot struct {
	CaRootCert
	Key []byte `json:"key"`
}

type ProvidedCa struct {
	Roots []CaRoot `json:"roots"`
}

type ProvidedCaManager interface {
	AddCaRoot(ctx context.Context, mesh string, root tls.KeyPair) error
	DeleteCaRoot(ctx context.Context, mesh string, id string) error

	DeleteCa(ctx context.Context, mesh string) error

	GetRootCerts(ctx context.Context, mesh string) ([]CaRootCert, error)
	GenerateWorkloadCert(ctx context.Context, mesh string, workload string) (*tls.KeyPair, error)
}

type providedCaManager struct {
	secretManager secret_manager.SecretManager
}

var _ ProvidedCaManager = &providedCaManager{}

func NewProvidedCaManager(secretManager secret_manager.SecretManager) ProvidedCaManager {
	return &providedCaManager{secretManager}
}

func (p *providedCaManager) AddCaRoot(ctx context.Context, mesh string, root tls.KeyPair) error {
	providedCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Get(ctx, providedCaSecret, core_store.GetBy(providedCaSecretKey(mesh))); err != nil {
		if core_store.IsResourceNotFound(err) {
			if err := p.secretManager.Create(ctx, providedCaSecret, core_store.CreateBy(providedCaSecretKey(mesh))); err != nil {
				return errors.Wrapf(err, "could not create CA for mesh %q", mesh)
			}
		} else {
			return errors.Wrapf(err, "failed to load CA for mesh %q", mesh)
		}
	}

	providedCa := ProvidedCa{}
	if len(providedCaSecret.Spec.Value) > 0 {
		if err := json.Unmarshal(providedCaSecret.Spec.Value, &providedCa); err != nil {
			return errors.Wrapf(err, "failed to deserialize a Root CA cert for Mesh %q", mesh)
		}
	}

	if len(providedCa.Roots) > 0 {
		return errors.New("cannot add more than 1 CA root to provided CA")
	}

	caRoot := CaRoot{
		Key: root.KeyPEM,
		CaRootCert: CaRootCert{
			Id:   uuid.New().String(),
			Cert: root.CertPEM,
		},
	}
	providedCa.Roots = append(providedCa.Roots, caRoot)

	caBytes, err := json.Marshal(providedCa)
	if err != nil {
		return errors.Wrap(err, "failed to marshal CA")
	}
	providedCaSecret.Spec.Value = caBytes
	if err := p.secretManager.Update(ctx, providedCaSecret); err != nil {
		return errors.Wrapf(err, "failed to update CA for mesh %q", mesh)
	}
	return nil
}

func (p *providedCaManager) DeleteCaRoot(ctx context.Context, mesh string, id string) error {
	meshCa, err := p.getMeshCa(ctx, mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to load CA key pair for Mesh %q", mesh)
	}

	var retainedCaRoots []CaRoot
	for _, root := range meshCa.Roots {
		if root.Id != id {
			retainedCaRoots = append(retainedCaRoots, root)
		}
	}

	if len(retainedCaRoots) == len(meshCa.Roots) {
		return errors.Errorf("could not find CA Root of id %q for mesh %q", id, mesh)
	}

	meshCa.Roots = retainedCaRoots
	newBytes, err := json.Marshal(meshCa)
	if err != nil {
		return err
	}

	providedCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Get(ctx, providedCaSecret, core_store.GetBy(providedCaSecretKey(mesh))); err != nil {
		return errors.Wrapf(err, "failed to load CA for mesh %q", mesh)
	}
	// todo(jakubdyszkiewicz) should we delete CA when there are 0 certs?
	providedCaSecret.Spec.Value = newBytes
	if err := p.secretManager.Update(ctx, providedCaSecret); err != nil {
		return errors.Wrapf(err, "failed to update CA for mesh %q", mesh)
	}
	return nil
}

func (p *providedCaManager) DeleteCa(ctx context.Context, mesh string) error {
	secretKey := providedCaSecretKey(mesh)
	builtinCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Delete(ctx, builtinCaSecret, core_store.DeleteBy(secretKey)); err != nil {
		if core_store.IsResourceNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "failed to delete Builtin CA for Mesh %q", mesh)
	}
	return nil
}

func (p *providedCaManager) GetRootCerts(ctx context.Context, mesh string) ([]CaRootCert, error) {
	meshCa, err := p.getMeshCa(ctx, mesh)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q", mesh)
	}
	caRootCerts := make([]CaRootCert, len(meshCa.Roots))
	for i, root := range meshCa.Roots {
		caRootCerts[i] = CaRootCert{
			Id:   root.Id,
			Cert: root.Cert,
		}
	}
	return caRootCerts, nil
}

func (p *providedCaManager) GenerateWorkloadCert(ctx context.Context, mesh string, workload string) (*tls.KeyPair, error) {
	meshCa, err := p.getMeshCa(ctx, mesh)
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

func (p *providedCaManager) getMeshCa(ctx context.Context, mesh string) (*ProvidedCa, error) {
	secretKey := providedCaSecretKey(mesh)
	providedCaSecret := &core_system.SecretResource{}
	if err := p.secretManager.Get(ctx, providedCaSecret, core_store.GetBy(secretKey)); err != nil {
		return nil, err
	}
	providedCa := ProvidedCa{}
	if err := json.Unmarshal(providedCaSecret.Spec.Value, &providedCa); err != nil {
		return nil, errors.Wrapf(err, "failed to deserialize a Root CA cert for Mesh %q", mesh)
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
