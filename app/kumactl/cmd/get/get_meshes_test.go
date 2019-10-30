package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"
)

var _ = Describe("kumactl get meshes", func() {

	sampleMeshes := []*mesh.MeshResource{
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: true,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Builtin_{
							Builtin: &v1alpha1.CertificateAuthority_Builtin{},
						},
					},
				},
				Logging: &v1alpha1.Logging{
					AccessLogs: &v1alpha1.Logging_AccessLogs{
						Enabled:  true,
						FilePath: "/tmp/access.log",
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "mesh1",
				Name:      "mesh1",
				Namespace: "",
			},
		},
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: false,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Builtin_{
							Builtin: &v1alpha1.CertificateAuthority_Builtin{},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "mesh2",
				Name:      "mesh2",
				Namespace: "",
			},
		},
	}

	Describe("GetMeshesCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: time.Now,
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, ds := range sampleMeshes {
				key := core_model.ResourceKey{
					Mesh:      ds.Meta.GetMesh(),
					Namespace: ds.Meta.GetNamespace(),
					Name:      ds.Meta.GetName(),
				}
				err := store.Create(context.Background(), ds, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get meshes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "meshes"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(buf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-meshes.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-meshes.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})

})
