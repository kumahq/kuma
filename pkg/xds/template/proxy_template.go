package template

import (
	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
)

const (
	ProfileDefaultProxy = "default-proxy"
)

var (
	DefaultProxyTemplate = &kuma_mesh.ProxyTemplate{
		Imports: []string{
			ProfileDefaultProxy,
		},
	}
)
