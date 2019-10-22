package yaml_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/yaml"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_rest "github.com/Kong/kuma/pkg/core/resources/model/rest"
)

var _ = Describe("printer", func() {

	var printer output.Printer
	var buf *bytes.Buffer

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

			// when
			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(buf.String()).To(Equal(string(expected)))
		},
		Entry("format `nil` value", testCase{
			obj:        nil,
			goldenFile: "nil.golden.yaml",
		}),
		Entry("format response from Kuma REST API", testCase{
			obj: &core_rest.Resource{
				Meta: core_rest.ResourceMeta{
					Type: string(mesh_core.MeshType),
					Name: "demo",
				},
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						Ca: &mesh_proto.CertificateAuthority{
							Type: &mesh_proto.CertificateAuthority_Builtin_{
								Builtin: &mesh_proto.CertificateAuthority_Builtin{},
							},
						},
						Enabled: true,
					},
				},
			},
			goldenFile: "mesh.golden.yaml",
		}),
	)
})
