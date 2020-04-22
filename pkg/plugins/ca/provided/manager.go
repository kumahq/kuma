package provided

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca"
	ca_issuer "github.com/Kong/kuma/pkg/core/ca/issuer"
	"github.com/Kong/kuma/pkg/core/datasource"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/plugins/ca/provided/config"
	"github.com/Kong/kuma/pkg/util/proto"
)

type providedCaManager struct {
	dataSourceLoader datasource.Loader
}

var _ ca.CaManager = &providedCaManager{}

func NewProvidedCaManager(dataSourceLoader datasource.Loader) ca.CaManager {
	return &providedCaManager{
		dataSourceLoader: dataSourceLoader,
	}
}

func (p *providedCaManager) ValidateBackend(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error {
	verr := validators.ValidationError{}

	cfg := &config.ProvidedCertificateAuthorityConfig{}
	if err := proto.ToTyped(backend.Config, cfg); err != nil {
		verr.AddViolation("", "could not convert backend config: "+err.Error())
		return verr.OrNil()
	}

	if cfg.GetCert() == nil {
		verr.AddViolation("cert", "has to be defined")
	} else {
		verr.AddError("cert", datasource.Validate(cfg.GetCert()))
	}
	if cfg.GetKey() == nil {
		verr.AddViolation("key", "has to be defined")
	} else {
		verr.AddError("key", datasource.Validate(cfg.GetKey()))
	}

	if !verr.HasViolations() {
		pair, err := p.getCa(ctx, mesh, backend)
		if err != nil {
			verr.AddViolation("cert", err.Error())
			verr.AddViolation("key", err.Error())
		} else {
			verr.AddError("", validateCaCert(pair))
		}
	}
	return verr.OrNil()
}

func (p *providedCaManager) getCa(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) (ca.KeyPair, error) {
	cfg := &config.ProvidedCertificateAuthorityConfig{}
	if err := proto.ToTyped(backend.Config, cfg); err != nil {
		return ca.KeyPair{}, errors.Wrap(err, "could not convert backend config to ProvidedCertificateAuthorityConfig")
	}
	key, err := p.dataSourceLoader.Load(ctx, mesh, cfg.Key)
	if err != nil {
		return ca.KeyPair{}, err
	}
	cert, err := p.dataSourceLoader.Load(ctx, mesh, cfg.Cert)
	if err != nil {
		return ca.KeyPair{}, err
	}
	pair := ca.KeyPair{
		CertPEM: cert,
		KeyPEM:  key,
	}
	return pair, nil
}

func (p *providedCaManager) Ensure(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error {
	return nil // Cert and Key are created by user and pointed in the configuration which is validated first
}

func (p *providedCaManager) GetRootCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) (ca.Cert, error) {
	meshCa, err := p.getCa(ctx, mesh, backend)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}
	return meshCa.CertPEM, nil
}

func (p *providedCaManager) GenerateDataplaneCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend, service string) (ca.KeyPair, error) {
	meshCa, err := p.getCa(ctx, mesh, backend)
	if err != nil {
		return ca.KeyPair{}, errors.Wrapf(err, "failed to load CA key pair for Mesh %q and backend %q", mesh, backend.Name)
	}

	keyPair, err := ca_issuer.NewWorkloadCert(meshCa, mesh, service)
	if err != nil {
		return ca.KeyPair{}, errors.Wrapf(err, "failed to generate a Workload Identity cert for workload %q in Mesh %q using backend %q", service, mesh, backend.Name)
	}
	return *keyPair, nil // todo pointer?
}
