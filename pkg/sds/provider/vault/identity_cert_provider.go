package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	vault "github.com/hashicorp/vault/api"

	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/provider"
)

func NewIdentityCertProvider(client *vault.Client) provider.SecretProvider {
	return &identityCertProvider{
		client: client,
	}
}

type identityCertProvider struct {
	client *vault.Client
}

var _ provider.SecretProvider = &identityCertProvider{}

func (m *identityCertProvider) Get(ctx context.Context, name string, requestor auth.Identity) (provider.Secret, error) {
	data := map[string]interface{}{
		"ttl":      "600h",
		"uri_sans": fmt.Sprintf("spiffe://%s/%s", requestor.Mesh, requestor.Service),
	}
	path := fmt.Sprintf("kuma-pki-%s/issue/dp-%s", requestor.Mesh, requestor.Service)
	secret, err := m.client.Logical().Write(path, data)
	if err != nil {
		switch tErr := err.(type) {
		case *vault.ResponseError:
			if tErr.StatusCode == 404 && len(tErr.Errors) == 1 && strings.Contains(tErr.Errors[0], "no handler for route") {
				return nil, errors.Errorf("there is no PKI enabled for %s mesh", requestor.Mesh)
			}
			if tErr.StatusCode == 400 && len(tErr.Errors) == 1 && tErr.Errors[0] == fmt.Sprintf("unknown role: dp-%s", requestor.Service) {
				return nil, errors.Errorf("there is no dp-%s role for PKI kuma-pki-%s", requestor.Service, requestor.Mesh)
			}
			if tErr.StatusCode == 403 && len(tErr.Errors) == 1 && tErr.Errors[0] == "permission denied" {
				return nil, errors.Errorf("permission denied - use token that allows to generate cert of %s", path)
			}
		}
		return nil, err
	}
	cert := secret.Data["certificate"].(string)
	key := secret.Data["private_key"].(string)

	return &provider.IdentityCertSecret{
		PemCerts: [][]byte{[]byte(cert)},
		PemKey:   []byte(key),
	}, nil
}

func (m *identityCertProvider) RequiresIdentity() bool {
	// Token from vault client is responsible for authz therefore we don't need to verify identity ourselves
	return false
}
