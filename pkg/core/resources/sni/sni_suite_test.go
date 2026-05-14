package sni_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
)

func TestSNI(t *testing.T) {
	registry.RegisterType(meshservice_api.MeshServiceResourceTypeDescriptor)
	registry.RegisterType(meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor)
	registry.RegisterType(meshmultizoneservice_api.MeshMultiZoneServiceResourceTypeDescriptor)
	RunSpecs(t, "SNI Compliance Warnings Suite")
}
