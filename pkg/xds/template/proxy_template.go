package template

import (
	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
)

var (
	DefaultProxyTemplate = &kuma_mesh.ProxyTemplate{
		Imports: []string{
			core_mesh.ProfileDefaultProxy,
		},
	}
)
