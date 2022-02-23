package get_test

import (
	"bytes"
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get traffic-traces", func() {

	trafficTraceResources := []*mesh.TrafficTraceResource{
		{
			Spec: &v1alpha1.TrafficTrace{
				Selectors: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web1",
							"version": "1.0",
						},
					},
				},
				Conf: &v1alpha1.TrafficTrace_Conf{
					Backend: "zipkin",
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "web1",
			},
		},
		{
			Spec: &v1alpha1.TrafficTrace{
				Selectors: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web2",
							"version": "1.0",
						},
					},
				},
				Conf: &v1alpha1.TrafficTrace_Conf{
					Backend: "zipkin",
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "web2",
			},
		},
	}

	Describe("GetTrafficTraceCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, mesh.TrafficTraceResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, ds := range trafficTraceResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
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

		DescribeTable("kumactl get traffic-traces -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "traffic-traces", given.outputFormat, given.pagination),
				).To(Succeed())

				// then
				Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-traffic-traces.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-traffic-traces.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				pagination:   "--size=1",
				goldenFile:   "get-traffic-traces.pagination.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-traffic-traces.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-traffic-traces.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})

})
