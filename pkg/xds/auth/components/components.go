package components

import (
	"context"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	"github.com/kumahq/kuma/pkg/xds/auth"
	k8s_auth "github.com/kumahq/kuma/pkg/xds/auth/k8s"
	universal_auth "github.com/kumahq/kuma/pkg/xds/auth/universal"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

type Deps struct {
	Config                  kuma_cp.Config
	Extensions              context.Context
	ReadOnlyResourceManager core_manager.ReadOnlyResourceManager
	XdsMetrics              *xds_metrics.Metrics
}

func NewKubeAuthenticator(deps Deps) (auth.Authenticator, error) {
	mgr, ok := k8s_extensions.FromManagerContext(deps.Extensions)
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	return k8s_auth.New(mgr.GetClient(), deps.XdsMetrics), nil
}

func NewUniversalAuthenticator(deps Deps) (auth.Authenticator, error) {
	config := deps.Config

	dataplaneValidator, err := builtin.NewDataplaneTokenValidator(deps.ReadOnlyResourceManager, config.Store.Type, config.DpServer.Authn.DpProxy.DpToken.Validator)
	if err != nil {
		return nil, err
	}
	zoneTokenValidator, err := builtin.NewZoneTokenValidator(deps.ReadOnlyResourceManager, config.IsFederatedZoneCP(), config.Store.Type, config.DpServer.Authn.ZoneProxy.ZoneToken.Validator)
	if err != nil {
		return nil, err
	}

	return universal_auth.NewAuthenticator(dataplaneValidator, zoneTokenValidator, config.Multizone.Zone.Name), nil
}

func DefaultAuthenticator(deps Deps, typ string) (auth.Authenticator, error) {
	switch typ {
	case dp_server.DpServerAuthServiceAccountToken:
		return NewKubeAuthenticator(deps)
	case dp_server.DpServerAuthDpToken, dp_server.DpServerAuthZoneToken:
		return NewUniversalAuthenticator(deps)
	case dp_server.DpServerAuthNone:
		return universal_auth.NewNoopAuthenticator(), nil
	default:
		return nil, errors.Errorf("unable to choose authenticator of %q", typ)
	}
}
