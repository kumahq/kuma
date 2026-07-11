package gateway

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/template"
)

var DefaultProxyTemplate = &mesh_proto.ProxyTemplate{
	Conf: &mesh_proto.ProxyTemplate_Conf{
		Imports: []string{
			metadata.ProfileGatewayProxy,
		},
	},
}

type TemplateResolver struct{}

var _ template.ProxyTemplateResolver = TemplateResolver{}

func (r TemplateResolver) GetTemplate(proxy *core_xds.Proxy) *mesh_proto.ProxyTemplate {
	if proxy.Dataplane == nil || !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	return DefaultProxyTemplate
}
