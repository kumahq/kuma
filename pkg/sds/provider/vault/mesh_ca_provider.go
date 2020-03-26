package vault

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	vault "github.com/hashicorp/vault/api"

	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/provider"
	"github.com/Kong/kuma/pkg/sds/provider/ca"
)

func NewMeshCaProvider(client *vault.Client) provider.SecretProvider {
	return &meshCaProvider{
		client: client,
	}
}

type meshCaProvider struct {
	client *vault.Client
}

var _ provider.SecretProvider = &meshCaProvider{}

func (m *meshCaProvider) Get(ctx context.Context, name string, requestor auth.Identity) (provider.Secret, error) {
	req := m.client.NewRequest("GET", fmt.Sprintf("/v1/kuma-pki-%s/ca/pem", requestor.Mesh))
	resp, err := m.client.RawRequest(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		switch tErr := err.(type) {
		case *vault.ResponseError:
			if tErr.StatusCode == 404 && len(tErr.Errors) == 1 && strings.Contains(tErr.Errors[0], "no handler for route") {
				return nil, errors.Errorf("there is no PKI enabled for %s mesh", requestor.Mesh)
			}
			if tErr.StatusCode == 403 && len(tErr.Errors) == 1 && tErr.Errors[0] == "permission denied" {
				return nil, errors.Errorf("permission denied - use token that allows to read CA cert of /v1/kuma-pki-%s/ca/pem", requestor.Mesh)
			}
		}
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// just in case, but should not happen since responses other than 200 are wrapped in client.RawRequest into api.ResponseError
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("error from Vault %d: %s", resp.StatusCode, body)
	}

	return &ca.MeshCaSecret{
		PemCerts: [][]byte{body},
	}, nil
}

func (m *meshCaProvider) RequiresIdentity() bool {
	return false
}
