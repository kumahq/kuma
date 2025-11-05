package samples

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
)

func MeshDefault() *mesh.MeshResource {
	return MeshDefaultBuilder().Build()
}

func MeshDefaultBuilder() *builders.MeshBuilder {
	return builders.Mesh()
}

func MeshMTLSBuilder() *builders.MeshBuilder {
	return MeshDefaultBuilder().
		WithEnabledMTLSBackend("builtin-1").
		AddBuiltinMTLSBackend("builtin-1")
}

func MeshMTLS() *mesh.MeshResource {
	return MeshMTLSBuilder().Build()
}
