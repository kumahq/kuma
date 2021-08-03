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

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/test"
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

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: func() time.Time { return rootTime },
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
					NewAPIServerClient: test.GetMockNewAPIServerClient(),
				},
			}

			store = core_store.NewPaginationStore(memory_resources.NewStore())

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
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get zone-ingresses -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "zone-ingresses"}, given.outputFormat, given.pagination))

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
				goldenFile:   "get-zone-ingresses.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-zone-ingresses.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-zone-ingresses.pagination.golden.txt",
				pagination:   "--size=1",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-zone-ingresses.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-zone-ingresses.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
