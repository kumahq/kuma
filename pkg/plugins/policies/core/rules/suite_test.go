package rules_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/test"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
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
