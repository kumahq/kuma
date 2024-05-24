package v1alpha1_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("MeshPassthrough", func() {
	Describe("Validate()", func() {
		type testCase struct {
			file string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshPassthrough := v1alpha1.NewMeshPassthroughResource()

				// when
				contents, err := os.ReadFile(path.Join("testdata", given.file+".input.yaml"))
				Expect(err).ToNot(HaveOccurred())
				err = core_model.FromYAML(contents, &meshPassthrough.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshPassthrough.Validate()
				actual, err := yaml.Marshal(verr)
				// have to do this otherwise valid cases will have null in the contents
				if string(actual) == "null\n" {
					actual = []byte{}
				}
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(matchers.MatchGoldenYAML(path.Join("testdata", given.file+".output.yaml")))
			},
			Entry("valid example", testCase{
				file: "valid",
			}),
			Entry("full failing example", testCase{
				file: "full-invalid",
			}),
		)
	})
})
