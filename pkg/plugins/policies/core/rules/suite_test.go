package rules_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/test"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	"github.com/kumahq/kuma/pkg/xds/context"
)

func TestRules(t *testing.T) {
	test.RunSpecs(t, "Rules Suite")
}

func readInputFile(inputFile string) []core_model.Resource {
	inputs, err := os.ReadFile(inputFile)
	Expect(err).NotTo(HaveOccurred())
	parts := strings.SplitN(string(inputs), "\n", 2)
	Expect(parts[0]).To(HavePrefix("#"), "is not a comment which explains the test")
	policiesBytesList := util_yaml.SplitYAML(string(inputs))

	var policies []core_model.Resource
	for _, policyBytes := range policiesBytesList {
		policy, err := rest.YAML.UnmarshalCore([]byte(policyBytes))
		Expect(err).ToNot(HaveOccurred())
		policies = append(policies, policy)
	}
	return policies
}

func buildMeshContext(rs []core_model.Resource) context.Resources {
	meshCtxResources := context.Resources{MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{}}
	for _, p := range rs {
		if _, ok := meshCtxResources.MeshLocalResources[p.Descriptor().Name]; !ok {
			meshCtxResources.MeshLocalResources[p.Descriptor().Name] = registry.Global().MustNewList(p.Descriptor().Name)
		}
		Expect(meshCtxResources.MeshLocalResources[p.Descriptor().Name].AddItem(p)).To(Succeed())
	}
	return meshCtxResources
}

func matchedPolicies(rs []core_model.Resource) []core_model.Resource {
	var matched []core_model.Resource
	for _, p := range rs {
		if strings.HasPrefix(p.GetMeta().GetName(), "matched-for-rules-") {
			matched = append(matched, p)
		}
	}
	return matched
}
