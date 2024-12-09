package runtime

import (
	"context"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
	"github.com/kumahq/kuma/pkg/xds/auth/components"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

type XDSRuntimeContext struct {
	DpProxyAuthenticator   xds_auth.Authenticator
	ZoneProxyAuthenticator xds_auth.Authenticator
	Hooks                  *xds_hooks.Hooks
	ServerCallbacks        util_xds.MultiXDSCallbacks
	Metrics                *xds_metrics.Metrics
}

type ContextWithXDS interface {
	Config() kuma_cp.Config
	Extensions() context.Context
	ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager
	Metrics() core_metrics.Metrics
	XDS() XDSRuntimeContext
}

func WithDefaults(ctx ContextWithXDS) (XDSRuntimeContext, error) {
	currentXDS := ctx.XDS()
	if currentXDS.Metrics == nil {
		xdsMetrics, err := xds_metrics.NewMetrics(ctx.Metrics())
		if err != nil {
			return XDSRuntimeContext{}, err
		}
		currentXDS.Metrics = xdsMetrics
	}
	authDeps := components.Deps{
		Config:                  ctx.Config(),
		Extensions:              ctx.Extensions(),
		ReadOnlyResourceManager: ctx.ReadOnlyResourceManager(),
		XdsMetrics:              currentXDS.Metrics,
	}

	if currentXDS.DpProxyAuthenticator == nil {
		dpProxyAuth, err := components.DefaultAuthenticator(authDeps, ctx.Config().DpServer.Authn.DpProxy.Type)
		if err != nil {
			return XDSRuntimeContext{}, err
		}
		currentXDS.DpProxyAuthenticator = dpProxyAuth
	}

	if currentXDS.ZoneProxyAuthenticator == nil {
		zoneProxyAuth, err := components.DefaultAuthenticator(authDeps, ctx.Config().DpServer.Authn.ZoneProxy.Type)
		if err != nil {
			return XDSRuntimeContext{}, err
		}
		currentXDS.ZoneProxyAuthenticator = zoneProxyAuth
	}

	if currentXDS.Hooks == nil {
		currentXDS.Hooks = &xds_hooks.Hooks{}
	}

	return currentXDS, nil
}

func (x XDSRuntimeContext) PerProxyTypeAuthenticator() xds_auth.Authenticator {
	return xds_auth.NewPerProxyTypeAuthenticator(x.DpProxyAuthenticator, x.ZoneProxyAuthenticator)
}
