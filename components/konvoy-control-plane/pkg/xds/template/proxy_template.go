package template

import (
	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
)

const (
	ProfileTransparentInboundProxy  = "transparent-inbound-proxy"
	ProfileTransparentOutboundProxy = "transparent-outbound-proxy"
)

var (
	TransparentProxyTemplate = &konvoy_mesh.ProxyTemplate{
		Sources: []*konvoy_mesh.ProxyTemplateSource{
			&konvoy_mesh.ProxyTemplateSource{
				Type: &konvoy_mesh.ProxyTemplateSource_Profile{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: ProfileTransparentInboundProxy,
					},
				},
			},
			&konvoy_mesh.ProxyTemplateSource{
				Type: &konvoy_mesh.ProxyTemplateSource_Profile{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: ProfileTransparentOutboundProxy,
					},
				},
			},
		},
	}
)
