package ca

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/tls"
)

type Cert = []byte

type KeyPair = tls.KeyPair

// Manager manages CAs by creating CAs and generating certificate. It is created per CA type and then may be used for different CA instances of the same type
type Manager interface {
	// ValidateBackend validates that backend configuration is correct
	ValidateBackend(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) error
	// Ensure ensures that CA of given name is available
	Ensure(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) error
	// UsedSecrets returns a list of secrets that are used by the manager
	UsedSecrets(mesh string, backend *mesh_proto.CertificateAuthorityBackend) ([]string, error)

	// GetRootCert returns root certificates of the CA
	GetRootCert(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend) ([]Cert, error)
	// GenerateDataplaneCert generates cert for a dataplane with service tags
	GenerateDataplaneCert(ctx context.Context, mesh string, backend *mesh_proto.CertificateAuthorityBackend, tags mesh_proto.MultiValueTagSet) (KeyPair, error)
}

// Managers hold Manager instance for each type of backend available (by default: builtin, provided)
type Managers = map[string]Manager
