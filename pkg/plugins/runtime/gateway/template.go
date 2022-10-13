package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/template"
)

// DefaultProxyTemplate captures the gateway proxy profile as a
// ProxyTemplate resource.
var DefaultProxyTemplate = &mesh_proto.ProxyTemplate{
	Conf: &mesh_proto.ProxyTemplate_Conf{
		Imports: []string{
			metadata.ProfileGatewayProxy,
		},
	},
}

// TemplateResolver resolved the default proxy template profile for
// builtin gateway dataplanes.
type TemplateResolver struct{}

var _ template.ProxyTemplateResolver = TemplateResolver{}

func (r TemplateResolver) GetTemplate(proxy *core_xds.Proxy) *mesh_proto.ProxyTemplate {
	if proxy.Dataplane == nil {
		return nil
	}

	if !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	// Return the builtin gateway template.
	return DefaultProxyTemplate
}
