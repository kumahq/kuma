package template

import (
	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
)

const (
	ProfileDefaultProxy = "default-proxy"
)

var (
	DefaultProxyTemplate = &kuma_mesh.ProxyTemplate{
		Conf: []*kuma_mesh.ProxyTemplateSource{
			&kuma_mesh.ProxyTemplateSource{
				Type: &kuma_mesh.ProxyTemplateSource_Profile{
					Profile: &kuma_mesh.ProxyTemplateProfileSource{
						Name: ProfileDefaultProxy,
					},
				},
			},
		},
	}
)
