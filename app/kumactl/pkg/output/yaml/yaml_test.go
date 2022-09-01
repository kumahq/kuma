package yaml_test

import (
	"bytes"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("printer", func() {

	var printer output.Printer
	var buf *bytes.Buffer
	t1, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
	t2, _ := time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")
	BeforeEach(func() {
		printer = yaml.NewPrinter()
		buf = &bytes.Buffer{}
	})

	type testCase struct {
		obj        interface{}
		goldenFile string
	}

	DescribeTable("should produce formatted output",
		func(given testCase) {
			// when
			err := printer.Print(given.obj, buf)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(buf.String()).To(MatchGoldenYAML(filepath.Join("testdata", given.goldenFile)))
		},
		Entry("format `nil` value", testCase{
			obj:        nil,
			goldenFile: "nil.golden.yaml",
		}),
		Entry("format response from Kuma REST API", testCase{
			obj: &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Type:             string(core_mesh.MeshType),
					Name:             "demo",
					CreationTime:     t1,
					ModificationTime: t2,
				},
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
						},
					},
				},
			},
			goldenFile: "mesh.golden.yaml",
		}),
	)
})
