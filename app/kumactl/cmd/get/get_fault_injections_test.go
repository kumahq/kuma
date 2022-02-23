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
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("kumactl get fault-injections", func() {

	faultInjectionResources := []*mesh.FaultInjectionResource{
		{
			Spec: &v1alpha1.FaultInjection{
				Sources: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "frontend",
							"version": "0.1",
						},
					},
				},
				Destinations: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
				Conf: &v1alpha1.FaultInjection_Conf{
					Delay: &v1alpha1.FaultInjection_Conf_Delay{
						Percentage: util_proto.Double(50),
						Value:      util_proto.Duration(time.Second * 5),
					},
					Abort: &v1alpha1.FaultInjection_Conf_Abort{
						Percentage: util_proto.Double(50),
						HttpStatus: util_proto.UInt32(500),
					},
					ResponseBandwidth: &v1alpha1.FaultInjection_Conf_ResponseBandwidth{
						Percentage: util_proto.Double(50),
						Limit:      util_proto.String("50 mbps"),
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "fi1",
			},
		},
		{
			Spec: &v1alpha1.FaultInjection{
				Sources: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web",
							"version": "0.1",
						},
					},
				},
				Destinations: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "redis",
						},
					},
				},
				Conf: &v1alpha1.FaultInjection_Conf{
					Delay: &v1alpha1.FaultInjection_Conf_Delay{
						Percentage: util_proto.Double(50),
						Value:      util_proto.Duration(time.Second * 5),
					},
					Abort: &v1alpha1.FaultInjection_Conf_Abort{
						Percentage: util_proto.Double(50),
						HttpStatus: util_proto.UInt32(500),
					},
					ResponseBandwidth: &v1alpha1.FaultInjection_Conf_ResponseBandwidth{
						Percentage: util_proto.Double(50),
						Limit:      util_proto.String("50 mbps"),
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "fi2",
			},
		},
	}

	Describe("GetFaultInjectionCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, mesh.FaultInjectionResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, ds := range faultInjectionResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			pagination   string
			matcher      func(path ...string) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get fault-injections -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "fault-injections", given.outputFormat, given.pagination),
				).To(Succeed())

				// then
				Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-fault-injections.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-fault-injections.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-fault-injections.pagination.golden.txt",
				pagination:   "--size=1",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-fault-injections.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-fault-injections.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
