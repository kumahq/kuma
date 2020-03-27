package sds

import (
	"context"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

// dpSdsHandler supports only one predefined dataplane since it's designed to run embedded SDS in Kuma DP
// The server should be protected so only the process of Kuma DP can use it
type dpSdsHandler struct {
	dpIdentity             sds_auth.Identity
	identitySecretProvider sds_provider.SecretProvider
	meshSecretProvider     sds_provider.SecretProvider
}

func (d *dpSdsHandler) Handle(ctx context.Context, req envoy.DiscoveryRequest) (*envoy_auth.Secret, error) {
	resource := req.ResourceNames[0]
	secretProvider, err := d.selectProvider(resource)
	if err != nil {
		return nil, err
	}

	secret, err := secretProvider.Get(ctx, resource, d.dpIdentity)
	if err != nil {
		return nil, err
	}
	return secret.ToResource(resource), nil
}

func (d *dpSdsHandler) selectProvider(resource string) (sds_provider.SecretProvider, error) {
	switch resource {
	case sds_provider.MeshCaResource:
		return d.meshSecretProvider, nil
	case sds_provider.IdentityCertResource:
		return d.identitySecretProvider, nil
	default:
		return nil, errors.Errorf("SDS request for %q resource is not supported", resource)
	}
}
