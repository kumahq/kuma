package get_test

import (
	"bytes"
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get dataplanes", func() {

	var dataplanes []*core_mesh.DataplaneResource
	BeforeEach(func() {
		// setup
		dataplanes = []*core_mesh.DataplaneResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "experiment",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 80,
								Tags: map[string]string{
									"service": "mobile",
									"version": "v1",
								},
							},
							{
								Port:        8090,
								ServicePort: 90,
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
					Mesh: "default",
					Name: "example",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.2",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 80,
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

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, core_mesh.DataplaneResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, pt := range dataplanes {
				key := core_model.ResourceKey{
					Mesh: pt.Meta.GetMesh(),
					Name: pt.Meta.GetName(),
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
			pagination   string
			goldenFile   string
			matcher      func(path ...string) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get dataplanes -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "dataplanes", given.outputFormat, given.pagination),
				).To(Succeed())

				// then
				Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-dataplanes.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-dataplanes.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-dataplanes.pagination.golden.txt",
				pagination:   "--size=1",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-dataplanes.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-dataplanes.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
