package samples

import (
	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
)

func ZipkinBackend() *meshtrace_proto.MeshTrace_ZipkinBackend {
	return &meshtrace_proto.MeshTrace_ZipkinBackend{
		Url: "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
	}
}
