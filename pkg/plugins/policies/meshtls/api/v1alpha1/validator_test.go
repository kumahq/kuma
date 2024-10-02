package v1alpha1_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("MeshTLS", func() {
	Describe("Validate()", func() {
		type testCase struct {
			name string
			file string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshTLS := v1alpha1.NewMeshTLSResource()

				// when
				contents, err := os.ReadFile(path.Join("testdata", given.file+".input.yaml"))
				Expect(err).ToNot(HaveOccurred())
				err = core_model.FromYAML(contents, &meshTLS.Spec)
				Expect(err).ToNot(HaveOccurred())

				meshTLS.SetMeta(&test_model.ResourceMeta{
					Name: given.name,
					Mesh: core_model.DefaultMesh,
				})
				// and
				verr := meshTLS.Validate()
				actual, err := yaml.Marshal(verr)
				// have to do this otherwise valid cases will have null in the contents
				if string(actual) == "null\n" {
					actual = []byte{}
				}
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(matchers.MatchGoldenYAML(path.Join("testdata", given.file+".output.yaml")))
			},
			Entry("full passing example", testCase{
				name: "meshtls-1",
				file: "full-valid",
			}),
			Entry("full failing example", testCase{
				name: "meshtls-2",
				file: "full-invalid",
			}),
			Entry("invalid top level", testCase{
				name: "meshtls-3",
				file: "invalid-top-level",
			}),
			Entry("full passing without top level", testCase{
				name: "meshtls-4",
				file: "full-valid-no-top-target",
			}),
		)
	})
})
