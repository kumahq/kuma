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

var _ = Describe("kumactl get zone-ingresses", func() {

	zoneIngresses := []*core_mesh.ZoneIngressResource{
		{
			Spec: &mesh_proto.ZoneIngress{
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address:           "1.1.1.1",
					Port:              10001,
					AdvertisedAddress: "2.2.2.2",
					AdvertisedPort:    20002,
				},
				AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
					{Mesh: "mesh-1", Tags: map[string]string{mesh_proto.ServiceTag: "svc-1"}},
					{Mesh: "mesh-2", Tags: map[string]string{mesh_proto.ServiceTag: "svc-2"}},
					{Mesh: "mesh-3", Tags: map[string]string{mesh_proto.ServiceTag: "svc-3"}},
				},
			},
			Meta: &test_model.ResourceMeta{
				Name: "ingress-zone-1",
			},
		},
		{
			Spec: &mesh_proto.ZoneIngress{
				Zone: "us-east",
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address:           "3.3.3.3",
					Port:              30003,
					AdvertisedAddress: "4.4.4.4",
					AdvertisedPort:    40004,
				},
				AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
					{Mesh: "mesh-3", Tags: map[string]string{mesh_proto.ServiceTag: "svc-2"}},
					{Mesh: "mesh-4", Tags: map[string]string{mesh_proto.ServiceTag: "svc-3"}},
					{Mesh: "mesh-5", Tags: map[string]string{mesh_proto.ServiceTag: "svc-4"}},
				},
			},
			Meta: &test_model.ResourceMeta{
				Name: "ingress-zone-2",
			},
		},
	}

	Describe("GetZoneIngressCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, core_mesh.ZoneIngressResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, cb := range zoneIngresses {
				err := store.Create(context.Background(), cb, core_store.CreateBy(core_model.MetaToResourceKey(cb.GetMeta())))
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

		DescribeTable("kumactl get zone-ingresses -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "zone-ingresses", given.outputFormat, given.pagination),
				).To(Succeed())

				// then
				Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-zone-ingresses.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-zone-ingresses.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-zone-ingresses.pagination.golden.txt",
				pagination:   "--size=1",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-zone-ingresses.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-zone-ingresses.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
