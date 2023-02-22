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

func Default(ctx components.Context) (XDSRuntimeContext, error) {
	auth, err := components.DefaultAuthenticator(ctx)
	if err != nil {
		return XDSRuntimeContext{}, err
	}

	return XDSRuntimeContext{
		Hooks:                  &xds_hooks.Hooks{},
		DpProxyAuthenticator:   auth,
		ZoneProxyAuthenticator: auth,
	}, nil
}

func (x XDSRuntimeContext) PerProxyTypeAuthenticator() xds_auth.Authenticator {
	return xds_auth.NewPerProxyTypeAuthenticator(x.DpProxyAuthenticator, x.ZoneProxyAuthenticator)
}
