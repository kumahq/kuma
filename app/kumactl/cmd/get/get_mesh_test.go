package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get mesh NAME", func() {
	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var outbuf, errbuf *bytes.Buffer
	var store core_store.ResourceStore
	var mesh *mesh_core.MeshResource
	BeforeEach(func() {
		// setup
		mesh = &mesh_core.MeshResource{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: true,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Builtin_{
							Builtin: &v1alpha1.CertificateAuthority_Builtin{},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh-1",
				Name: "mesh-1",
			},
		}
		key := core_model.ResourceKey{
			Mesh: mesh.Meta.GetMesh(),
			Name: mesh.Meta.GetName(),
		}
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Now: time.Now,
				NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
					return store, nil
				},
			},
		}
		store = memory_resources.NewStore()
		err := store.Create(context.Background(), mesh, core_store.CreateBy(key))
		Expect(err).ToNot(HaveOccurred())

		rootCmd = cmd.NewRootCmd(rootCtx)
		outbuf = &bytes.Buffer{}
		errbuf = &bytes.Buffer{}
		rootCmd.SetOut(outbuf)
		rootCmd.SetErr(errbuf)
	})
	It("should throw an error in case of no args", func() {
		// given
		rootCmd.SetArgs([]string{
			"get", "mesh"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("requires at least 1 arg(s), only received 0"))
		// and
		Expect(outbuf.String()).To(MatchRegexp(`Error: requires at least 1 arg\(s\), only received 0`))
		// and
		Expect(errbuf.Bytes()).To(BeEmpty())
	})
	It("should return error message if doesn't exist", func() {
		//given
		rootCmd.SetArgs([]string{
			"get", "mesh", "mesh-2"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(outbuf.String()).To(Equal("Error: No resources found in mesh-2 mesh\n"))
		// and
		Expect(errbuf.Bytes()).To(BeEmpty())

	})
	Describe("kumactl get mesh NAME -o table|json|yaml", func() {

		type testCase struct {
			outputFormat string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get dataplanes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"get", "mesh", "mesh-1"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(outbuf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-mesh.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
		)
	})
})
