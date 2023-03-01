package components

import (
	"context"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
	"github.com/kumahq/kuma/pkg/xds/auth"
	k8s_auth "github.com/kumahq/kuma/pkg/xds/auth/k8s"
	universal_auth "github.com/kumahq/kuma/pkg/xds/auth/universal"
)

type Context interface {
	Config() kuma_cp.Config
	Extensions() context.Context
	ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager
}

func NewKubeAuthenticator(rt Context) (auth.Authenticator, error) {
	mgr, ok := k8s_extensions.FromManagerContext(rt.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	return k8s_auth.New(mgr.GetClient()), nil
}

func NewUniversalAuthenticator(rt Context) (auth.Authenticator, error) {
	config := rt.Config()

	dataplaneValidator, err := builtin.NewDataplaneTokenValidator(rt.ReadOnlyResourceManager(), config.Store.Type, config.DpServer.Authn.DpProxy.DpToken.Validator)
	if err != nil {
		return nil, err
	}
	zoneIngressValidator := builtin.NewZoneIngressTokenValidator(rt.ReadOnlyResourceManager(), config.Store.Type)
	zoneTokenValidator, err := builtin.NewZoneTokenValidator(rt.ReadOnlyResourceManager(), config.Mode, config.Store.Type, config.DpServer.Authn.ZoneProxy.ZoneToken.Validator)
	if err != nil {
		return nil, err
	}
	adaptedValidator := zoneingress.NewZoneValidatorAdapter(zoneIngressValidator, zoneTokenValidator)

	return universal_auth.NewAuthenticator(dataplaneValidator, adaptedValidator, config.Multizone.Zone.Name), nil
}

func DefaultAuthenticator(rt Context, typ string) (auth.Authenticator, error) {
	switch typ {
	case dp_server.DpServerAuthServiceAccountToken:
		return NewKubeAuthenticator(rt)
	case dp_server.DpServerAuthDpToken, dp_server.DpServerAuthZoneToken:
		return NewUniversalAuthenticator(rt)
	case dp_server.DpServerAuthNone:
		return universal_auth.NewNoopAuthenticator(), nil
	default:
		return nil, errors.Errorf("unable to choose authenticator of %q", typ)
	}
}
