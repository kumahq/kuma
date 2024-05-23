package v1alpha1_test

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("MeshExternalService", func() {
	Describe("Validate()", func() {
		type testCase struct {
			file string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				MeshExternalService := v1alpha1.NewMeshExternalServiceResource()

				// when
				contents, err := os.ReadFile(path.Join("testdata", given.file + "input.yaml"))
				Expect(err).ToNot(HaveOccurred())
				err = core_model.FromYAML(contents, &MeshExternalService.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := MeshExternalService.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(matchers.MatchGoldenYAML(given.file + "output.yaml"))
			},
			Entry("empty 'from' and 'to' array", testCase{
				file: "minimal-valid",
			}),
		)
	})
})
