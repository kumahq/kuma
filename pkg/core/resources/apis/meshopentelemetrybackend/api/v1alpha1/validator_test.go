package v1alpha1_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

var _ = Describe("MeshOpenTelemetryBackend", func() {
	Describe("Validate()", func() {
		type testCase struct {
			file string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				resource := v1alpha1.NewMeshOpenTelemetryBackendResource()

				contents, err := os.ReadFile(path.Join("testdata", given.file+".input.yaml"))
				Expect(err).ToNot(HaveOccurred())
				err = core_model.FromYAML(contents, &resource.Spec)
				Expect(err).ToNot(HaveOccurred())

				resource.SetMeta(&test_model.ResourceMeta{
					Name: "backend",
					Mesh: core_model.DefaultMesh,
				})

				verr := resource.Validate()
				actual, err := yaml.Marshal(verr)
				if string(actual) == "null\n" {
					actual = []byte{}
				}
				Expect(err).ToNot(HaveOccurred())

				Expect(actual).To(matchers.MatchGoldenYAML(path.Join("testdata", given.file+".output.yaml")))
			},
			Entry("minimal valid gRPC", testCase{
				file: "minimal-valid",
			}),
			Entry("full valid HTTP with path", testCase{
				file: "full-valid-http",
			}),
			Entry("gRPC with path is valid (path silently ignored)", testCase{
				file: "grpc-with-path-valid",
			}),
			Entry("missing address", testCase{
				file: "missing-address-invalid",
			}),
			Entry("missing port", testCase{
				file: "missing-port-invalid",
			}),
			Entry("invalid address", testCase{
				file: "invalid-address",
			}),
			Entry("port out of range", testCase{
				file: "port-out-of-range-invalid",
			}),
			Entry("invalid protocol", testCase{
				file: "invalid-protocol",
			}),
			Entry("path without leading slash", testCase{
				file: "path-no-slash-invalid",
			}),
			Entry("path with query string", testCase{
				file: "path-with-query-invalid",
			}),
			Entry("nodeEndpoint valid", testCase{
				file: "node-endpoint-valid",
			}),
			Entry("nodeEndpoint with path valid", testCase{
				file: "node-endpoint-with-path-valid",
			}),
			Entry("nodeEndpoint port out of range", testCase{
				file: "node-endpoint-port-invalid",
			}),
			Entry("both endpoint and nodeEndpoint set", testCase{
				file: "both-endpoints-invalid",
			}),
			Entry("neither endpoint nor nodeEndpoint set", testCase{
				file: "no-endpoint-invalid",
			}),
		)
	})
})
