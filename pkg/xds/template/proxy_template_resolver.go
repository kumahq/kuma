package template

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
)

type ProxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate
}

type StaticProxyTemplateResolver struct {
	Template *mesh_proto.ProxyTemplate
}

func (r *StaticProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	return r.Template
}
