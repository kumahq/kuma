package server

import (
	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
)

var (
	TransparentProxyTemplate = &konvoy_mesh.ProxyTemplate{
		Spec: konvoy_mesh.ProxyTemplateSpec{
			Sources: []konvoy_mesh.ProxyTemplateSource{
				{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: generator.ProfileTransparentInboundProxy,
					},
				},
				{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: generator.ProfileTransparentOutboundProxy,
					},
				},
			},
		},
	}
)

type proxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *konvoy_mesh.ProxyTemplate
}

type simpleProxyTemplateResolver struct {
	DefaultProxyTemplate *konvoy_mesh.ProxyTemplate
}

func (r *simpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *konvoy_mesh.ProxyTemplate {
	return r.DefaultProxyTemplate
}
