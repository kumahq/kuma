package v1alpha1_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("MeshExternalService", func() {
	Describe("Validate()", func() {
		type testCase struct {
			name string
			file string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshExternalService := v1alpha1.NewMeshExternalServiceResource()

				// when
				contents, err := os.ReadFile(path.Join("testdata", given.file+".input.yaml"))
				Expect(err).ToNot(HaveOccurred())
				err = core_model.FromYAML(contents, &meshExternalService.Spec)
				Expect(err).ToNot(HaveOccurred())

				meshExternalService.SetMeta(&test_model.ResourceMeta{
					Name: given.name,
					Mesh: core_model.DefaultMesh,
				})
				// and
				verr := meshExternalService.Validate()
				actual, err := yaml.Marshal(verr)
				// have to do this otherwise valid cases will have null in the contents
				if string(actual) == "null\n" {
					actual = []byte{}
				}
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(matchers.MatchGoldenYAML(path.Join("testdata", given.file+".output.yaml")))
			},
			Entry("minimal passing example", testCase{
				name: "external-service",
				file: "minimal-valid",
			}),
			Entry("full example without extension", testCase{
				name: "external-service",
				file: "full-without-extension-valid",
			}),
			Entry("minimal failing example with unknown port", testCase{
				name: "external-service",
				file: "minimal-invalid",
			}),
			Entry("failing example with missing extension type", testCase{
				name: "external-service",
				file: "minimal-invalid-extension",
			}),
			Entry("missing client-cert", testCase{
				name: "external-service",
				file: "missing-client-cert-invalid",
			}),
			Entry("full failing example", testCase{
				name: "external-service",
				file: "full-invalid",
			}),
			Entry("min tls version higher than max", testCase{
				name: "external-service",
				file: "min-higher-than-max-invalid",
			}),
			Entry("name too long", testCase{
				name: "external-service-very-long-very-long-very-long-very-long-very-long-very-long-very-long",
				file: "name-too-long",
			}),
			Entry("name length 63", testCase{
				name: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				file: "name-length-63",
			}),
		)
	})
})
