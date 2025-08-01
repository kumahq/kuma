package v1alpha1_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("MeshTrust", func() {
	DescribeTable("should validate all folders", func(inputFile string) {
		// setup
		meshTrust := v1alpha1.NewMeshTrustResource()

		// when
		contents, err := os.ReadFile(inputFile)
		Expect(err).ToNot(HaveOccurred())
		err = core_model.FromYAML(contents, &meshTrust.Spec)
		Expect(err).ToNot(HaveOccurred())

		meshTrust.SetMeta(&test_model.ResourceMeta{
			Name: "test",
			Mesh: core_model.DefaultMesh,
		})

		// and
		verr := meshTrust.Validate()
		actual, err := yaml.Marshal(verr)
		if string(actual) == "null\n" {
			actual = []byte{}
		}
		Expect(err).ToNot(HaveOccurred())

		// then
		goldenFile := strings.ReplaceAll(inputFile, ".input.yaml", ".golden.yaml")
		Expect(actual).To(matchers.MatchGoldenYAML(goldenFile))
	}, test.EntriesForFolder("spec"))
})
