package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	memory_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"

	test_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
)

var _ = Describe("kumactl get dataplanes", func() {

	var dataplanes []*mesh_core.DataplaneResource

	BeforeEach(func() {
		dataplanes = []*mesh_core.DataplaneResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh:      "default",
					Namespace: "trial",
					Name:      "experiment",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.1:8080:80",
								Tags: map[string]string{
									"service": "mobile",
									"version": "v1",
								},
							},
							{
								Interface: "127.0.0.1:8090:90",
								Tags: map[string]string{
									"service": "metrics",
									"version": "v1",
								},
							},
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh:      "default",
					Namespace: "demo",
					Name:      "example",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.2:8080:80",
								Tags: map[string]string{
									"service": "web",
									"version": "v2",
								},
							},
						},
					},
				},
			},
		}
	})

	Describe("GetDataplanesCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup

			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, pt := range dataplanes {
				key := core_model.ResourceKey{
					Mesh:      pt.Meta.GetMesh(),
					Namespace: pt.Meta.GetNamespace(),
					Name:      pt.Meta.GetName(),
				}
				err := store.Create(context.Background(), pt, core_store.CreateBy(key))
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

		DescribeTable("kumactl get dataplanes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "dataplanes"}, given.outputFormat))

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
				goldenFile:   "get-dataplanes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-dataplanes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-dataplanes.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-dataplanes.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
