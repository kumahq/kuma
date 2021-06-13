package components

import (
	"github.com/pkg/errors"
	kube_auth "k8s.io/api/authentication/v1"

	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	"github.com/kumahq/kuma/pkg/xds/auth"
	k8s_auth "github.com/kumahq/kuma/pkg/xds/auth/k8s"
	universal_auth "github.com/kumahq/kuma/pkg/xds/auth/universal"
)

func NewKubeAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	mgr, ok := k8s_extensions.FromManagerContext(rt.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_auth.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_auth.SchemeGroupVersion)
	}
	return k8s_auth.New(mgr.GetClient()), nil
}

func NewUniversalAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	issuer, err := builtin.NewDataplaneTokenIssuer(rt.ReadOnlyResourceManager())
	if err != nil {
		return nil, err
	}
	zoneIngressIssuer, err := builtin.NewZoneIngressTokenIssuer(rt.ReadOnlyResourceManager())
	if err != nil {
		return nil, err
	}
	return universal_auth.NewAuthenticator(issuer, zoneIngressIssuer, rt.Config().Multizone.Zone.Name), nil
}

func DefaultAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	switch rt.Config().DpServer.Auth.Type {
	case dp_server.DpServerAuthServiceAccountToken:
		return NewKubeAuthenticator(rt)
	case dp_server.DpServerAuthDpToken:
		return NewUniversalAuthenticator(rt)
	case dp_server.DpServerAuthNone:
		return universal_auth.NewNoopAuthenticator(), nil
	default:
		return nil, errors.Errorf("unable to choose authenticator of %q", rt.Config().DpServer.Auth.Type)
	}
}
