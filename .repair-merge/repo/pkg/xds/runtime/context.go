package runtime

import (
	"context"

	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	util_xds "github.com/kumahq/kuma/v3/pkg/util/xds"
	xds_auth "github.com/kumahq/kuma/v3/pkg/xds/auth"
	"github.com/kumahq/kuma/v3/pkg/xds/auth/components"
	xds_hooks "github.com/kumahq/kuma/v3/pkg/xds/hooks"
	xds_metrics "github.com/kumahq/kuma/v3/pkg/xds/metrics"
)

type XDSRuntimeContext struct {
	Authenticator   xds_auth.Authenticator
	Hooks           *xds_hooks.Hooks
	ServerCallbacks util_xds.MultiXDSCallbacks
	Metrics         *xds_metrics.Metrics
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

	if currentXDS.Authenticator == nil {
		authenticator, err := components.DefaultAuthenticator(authDeps, ctx.Config().DpServer.Authn.DpProxy.Type)
		if err != nil {
			return XDSRuntimeContext{}, err
		}
		currentXDS.Authenticator = authenticator
	}

	if currentXDS.Hooks == nil {
		currentXDS.Hooks = &xds_hooks.Hooks{}
	}

	return currentXDS, nil
}
