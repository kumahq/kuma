package matchers_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/test"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func TestMatchers(t *testing.T) {
	test.RunSpecs(t, "Matchers Suite")
}

func readPolicies(file string) (xds_context.Resources, []core_model.ResourceType) {
	responseBytes, err := os.ReadFile(file)
	Expect(err).ToNot(HaveOccurred())

	rawResources := util_yaml.SplitYAML(string(responseBytes))
	meshResources := map[core_model.ResourceType]core_model.ResourceList{}
	typesMap := map[core_model.ResourceType]struct{}{}

	for _, rawResource := range rawResources {
		resource, err := rest.YAML.UnmarshalCore([]byte(rawResource))
		Expect(err).ToNot(HaveOccurred())

		rType := resource.Descriptor().Name

		_, ok := meshResources[rType]
		if !ok {
			meshResources[rType] = resource.Descriptor().NewList()
		}
		Expect(meshResources[rType].AddItem(resource)).To(Succeed())

		typesMap[rType] = struct{}{}
	}

	resTypes := []core_model.ResourceType{}
	for rType := range typesMap {
		resTypes = append(resTypes, rType)
	}

	return xds_context.Resources{
		MeshLocalResources: meshResources,
	}, resTypes
}

func readDPP(file string) *core_mesh.DataplaneResource {
	dppYaml, err := os.ReadFile(file)
	Expect(err).ToNot(HaveOccurred())

	dpp, err := rest.YAML.UnmarshalCore(dppYaml)
	Expect(err).ToNot(HaveOccurred())
	return dpp.(*core_mesh.DataplaneResource)
}

func readES(file string) *core_mesh.ExternalServiceResource {
	dppYaml, err := os.ReadFile(file)
	Expect(err).ToNot(HaveOccurred())

	dpp, err := rest.YAML.UnmarshalCore(dppYaml)
	Expect(err).ToNot(HaveOccurred())
	return dpp.(*core_mesh.ExternalServiceResource)
}

func readMES(file string) *meshexternalservice_api.MeshExternalServiceResource {
	mesYaml, err := os.ReadFile(file)
	Expect(err).ToNot(HaveOccurred())

	mes, err := rest.YAML.UnmarshalCore(mesYaml)
	Expect(err).ToNot(HaveOccurred())
	return mes.(*meshexternalservice_api.MeshExternalServiceResource)
}
