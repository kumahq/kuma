package server

import (
	"context"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/pkg/errors"
	kube_auth "k8s.io/api/authentication/v1"

	config_core "github.com/Kong/kuma/pkg/config/core"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	k8s_sds_auth "github.com/Kong/kuma/pkg/sds/auth/k8s"
	universal_sds_auth "github.com/Kong/kuma/pkg/sds/auth/universal"
	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
	ca_sds_provider "github.com/Kong/kuma/pkg/sds/provider/ca"
	identity_sds_provider "github.com/Kong/kuma/pkg/sds/provider/identity"
	"github.com/Kong/kuma/pkg/tokens/builtin"
)

const (
	MeshCaResource       = "mesh_ca"
	IdentityCertResource = "identity_cert"
)

func NewKubeAuthenticator(rt core_runtime.Runtime) (sds_auth.Authenticator, error) {
	mgr, ok := k8s_runtime.FromManagerContext(rt.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_auth.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_auth.SchemeGroupVersion)
	}
	return k8s_sds_auth.New(mgr.GetClient(), DefaultDataplaneResolver(rt.ResourceManager())), nil
}

func NewUniversalAuthenticator(rt core_runtime.Runtime) (sds_auth.Authenticator, error) {
	dpResolver := DefaultDataplaneResolver(rt.ResourceManager())
	if !rt.Config().DataplaneTokenServer.Enabled {
		return universal_sds_auth.NewNoopAuthenticator(dpResolver), nil
	}
	issuer, err := builtin.NewDataplaneTokenIssuer(rt)
	if err != nil {
		return nil, err
	}
	return universal_sds_auth.NewAuthenticator(issuer, dpResolver), nil
}

func DefaultAuthenticator(rt core_runtime.Runtime) (sds_auth.Authenticator, error) {
	switch env := rt.Config().Environment; env {
	case config_core.KubernetesEnvironment:
		return NewKubeAuthenticator(rt)
	case config_core.UniversalEnvironment:
		return NewUniversalAuthenticator(rt)
	default:
		return nil, errors.Errorf("unable to choose SDS authenticator for environment type %q", env)
	}
}

func DefaultDataplaneResolver(resourceManager core_manager.ResourceManager) func(context.Context, core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
	return func(ctx context.Context, proxyId core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
		dataplane := &core_mesh.DataplaneResource{}
		if err := resourceManager.Get(ctx, dataplane, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
			return nil, err
		}
		return dataplane, nil
	}
}

func DefaultMeshCaProvider(rt core_runtime.Runtime) sds_provider.SecretProvider {
	return ca_sds_provider.New(rt.ResourceManager(), rt.CaManagers())
}

func DefaultIdentityCertProvider(rt core_runtime.Runtime) sds_provider.SecretProvider {
	return identity_sds_provider.New(rt.ResourceManager(), rt.CaManagers())
}

func DefaultSecretProviderSelector(rt core_runtime.Runtime) func(string) (sds_provider.SecretProvider, error) {
	meshCaProvider := DefaultMeshCaProvider(rt)
	identityCertProvider := DefaultIdentityCertProvider(rt)
	return func(resource string) (sds_provider.SecretProvider, error) {
		switch resource {
		case MeshCaResource:
			return meshCaProvider, nil
		case IdentityCertResource:
			return identityCertProvider, nil
		default:
			return nil, errors.Errorf("SDS request for %q resource is not supported", resource)
		}
	}
}

func DefaultSecretDiscoveryHandler(rt core_runtime.Runtime) (SecretDiscoveryHandler, error) {
	authenticator, err := DefaultAuthenticator(rt)
	if err != nil {
		return nil, err
	}
	secretProviderSelector := DefaultSecretProviderSelector(rt)
	return SecretDiscoveryHandlerFunc(func(ctx context.Context, req envoy.DiscoveryRequest) (*envoy_auth.Secret, error) {
		resource := req.ResourceNames[0]
		provider, err := secretProviderSelector(resource)
		if err != nil {
			return nil, err
		}
		proxyId, err := core_xds.ParseProxyId(req.Node)
		if err != nil {
			return nil, errors.Wrap(err, "SDS request must have a valid Proxy Id")
		}
		requestor := sds_auth.Identity{Mesh: proxyId.Mesh}
		if provider.RequiresIdentity() {
			credential, err := sds_auth.ExtractCredential(ctx)
			if err != nil {
				return nil, err
			}
			requestor, err = authenticator.Authenticate(ctx, *proxyId, credential)
			if err != nil {
				return nil, err
			}
		}
		secret, err := provider.Get(ctx, resource, requestor)
		if err != nil {
			return nil, err
		}
		return secret.ToResource(resource), nil
	}), nil
}

type SecretDiscoveryHandlerFunc func(ctx context.Context, req envoy.DiscoveryRequest) (*envoy_auth.Secret, error)

func (f SecretDiscoveryHandlerFunc) Handle(ctx context.Context, req envoy.DiscoveryRequest) (*envoy_auth.Secret, error) {
	return f(ctx, req)
}
