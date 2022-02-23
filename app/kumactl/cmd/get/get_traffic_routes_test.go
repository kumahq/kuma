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

var _ = Describe("kumactl get traffic-routes", func() {

	var sampleTrafficRoutes []*core_mesh.TrafficRouteResource

	BeforeEach(func() {
		sampleTrafficRoutes = []*core_mesh.TrafficRouteResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "web-to-backend",
				},
				Spec: &mesh_proto.TrafficRoute{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "backend-to-db",
				},
				Spec: &mesh_proto.TrafficRoute{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "gateway-to-service",
				},
				Spec: &mesh_proto.TrafficRoute{},
			},
		}
	})

	Describe("GetTrafficRoutesCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, core_mesh.TrafficRouteResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, pt := range sampleTrafficRoutes {
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

		DescribeTable("kumactl get traffic-routes -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "traffic-routes", given.outputFormat, given.pagination),
				).To(Succeed())

				// then
				Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-traffic-routes.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-traffic-routes.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				pagination:   "--size=1",
				goldenFile:   "get-traffic-routes.pagination.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-traffic-routes.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-traffic-routes.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
