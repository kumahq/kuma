package v1alpha1

import (
	"github.com/kumahq/kuma/v3/pkg/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("DoNothingPolicy")
)

type plugin struct{}

func (p *plugin) Order() int {
	panic("unimplemented")
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	log.Info("apply is not implemented")
	return nil
}
