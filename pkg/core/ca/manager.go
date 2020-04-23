package ca

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/tls"
)

type Cert = []byte

type KeyPair = tls.KeyPair

// Manager manages CAs by creating CAs and generating certificate. It is created per CA type and then may be used for different CA instances of the same type
type Manager interface {
	// Validates that backend configuration is correct
	ValidateBackend(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error
	// Ensure that CA of given name is created
	Ensure(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) error

	// Returns root certificates of the CA
	GetRootCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend) (Cert, error) // todo list of certs?
	// Generates cert for a dataplanes with service tag
	GenerateDataplaneCert(ctx context.Context, mesh string, backend mesh_proto.CertificateAuthorityBackend, service string) (KeyPair, error)
}

// Managers hold Manager instance for each type of backend available (by default: builtin, provided)
type Managers = map[string]Manager
