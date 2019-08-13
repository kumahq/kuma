package get_test

import (
	"bytes"
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	memory_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	test_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"

	pkg_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

var _ = Describe("konvoy get meshes", func() {

	sampleMeshes := []*mesh.MeshResource{
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Embedded_{
							Embedded: &v1alpha1.CertificateAuthority_Embedded{},
						},
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
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Embedded_{
							Embedded: &v1alpha1.CertificateAuthority_Embedded{},
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

		var rootCtx *pkg_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			rootCtx = &pkg_cmd.RootContext{
				Runtime: pkg_cmd.RootRuntime{
					Now: func() time.Time { return time.Now() },
					NewResourceStore: func(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
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

		DescribeTable("konvoyctl get meshes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
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
