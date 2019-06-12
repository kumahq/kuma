package template

import (
	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
)

const (
	ProfileTransparentInboundProxy  = "transparent-inbound-proxy"
	ProfileTransparentOutboundProxy = "transparent-outbound-proxy"
)

var (
	TransparentProxyTemplate = &konvoy_mesh.ProxyTemplate{
		Spec: konvoy_mesh.ProxyTemplateSpec{
			Sources: []konvoy_mesh.ProxyTemplateSource{
				{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: ProfileTransparentInboundProxy,
					},
				},
				{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: ProfileTransparentOutboundProxy,
					},
				},
			},
		},
	}
)
