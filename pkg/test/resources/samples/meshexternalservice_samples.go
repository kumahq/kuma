package samples

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func MeshExternalServiceExampleBuilder() *builders.MeshExternalServiceBuilder {
	return builders.MeshExternalService().
		WithKumaVIP("242.0.0.1")
}

func MeshExternalServiceExample() *v1alpha1.MeshExternalServiceResource {
	return MeshExternalServiceExampleBuilder().Build()
}
