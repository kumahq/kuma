package v1alpha1

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.CoreResourcePlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.CoreResourcePlugin {
	return &plugin{}
}

func (p *plugin) Generate(rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.WorkloadIdentity == nil {
		return nil
	}
	rs.AddSet(proxy.WorkloadIdentity.AdditionalResources)
	return nil
}
