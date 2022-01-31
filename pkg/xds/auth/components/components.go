package components

import (
	"github.com/pkg/errors"

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
	return k8s_auth.New(mgr.GetClient()), nil
}

func NewUniversalAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	config := rt.Config()

	dataplaneValidator := builtin.NewDataplaneTokenValidator(rt.ResourceManager())
	zoneIngressValidator := builtin.NewZoneIngressTokenValidator(rt.ResourceManager())
	zoneTokenValidator := builtin.NewZoneTokenValidator(rt.ResourceManager(), config.Mode)

	return universal_auth.NewAuthenticator(dataplaneValidator, zoneIngressValidator, zoneTokenValidator, config.Multizone.Zone.Name), nil
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
