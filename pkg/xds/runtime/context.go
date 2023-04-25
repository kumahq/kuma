package runtime

import (
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
	"github.com/kumahq/kuma/pkg/xds/auth/components"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
)

type XDSRuntimeContext struct {
	DpProxyAuthenticator   xds_auth.Authenticator
	ZoneProxyAuthenticator xds_auth.Authenticator
	Hooks                  *xds_hooks.Hooks
	ServerCallbacks        util_xds.Callbacks
}

type ContextWithXDS interface {
	components.Context
	XDS() XDSRuntimeContext
}

func WithDefaults(ctx ContextWithXDS) (XDSRuntimeContext, error) {
	currentXDS := ctx.XDS()

	if currentXDS.DpProxyAuthenticator == nil {
		dpProxyAuth, err := components.DefaultAuthenticator(ctx, ctx.Config().DpServer.Authn.DpProxy.Type)
		if err != nil {
			return XDSRuntimeContext{}, err
		}
		currentXDS.DpProxyAuthenticator = dpProxyAuth
	}

	if currentXDS.ZoneProxyAuthenticator == nil {
		zoneProxyAuth, err := components.DefaultAuthenticator(ctx, ctx.Config().DpServer.Authn.ZoneProxy.Type)
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
