package ca

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/tls"
)

type Cert = []byte

type KeyPair = tls.KeyPair

type CaManager interface { // todo drop prefix?
	ValidateBackend(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error
	Ensure(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error

	GetRootCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) (Cert, error) // todo list of certs?
	GenerateDataplaneCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend, service string) (KeyPair, error)
}

// CaManagers hold CaManager instance for each type of backend available (by default: builtin, provided)
type CaManagers = map[string]CaManager
