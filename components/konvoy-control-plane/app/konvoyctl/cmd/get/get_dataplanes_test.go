package get_test

import (
	"bytes"
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/get"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	memory_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"

	test_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
)

var _ = Describe("konvoy get dataplanes", func() {

	var now, t1, t2 time.Time
	var sampleDataplaneInsights []*mesh_core.DataplaneInsightResource

	BeforeEach(func() {
		now, _ = time.Parse(time.RFC3339, "2019-07-17T18:08:41+00:00")
		t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		t2, _ = time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")

		sampleDataplaneInsights = []*mesh_core.DataplaneInsightResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh:      "default",
					Namespace: "trial",
					Name:      "experiment",
				},
				Spec: mesh_proto.DataplaneInsight{
					Subscriptions: []*mesh_proto.DiscoverySubscription{
						{
							Id:                     "1",
							ControlPlaneInstanceId: "node-001",
							ConnectTime:            util_proto.MustTimestampProto(t1),
							Status: mesh_proto.DiscoverySubscriptionStatus{
								Total: mesh_proto.DiscoveryServiceStats{
									ResponsesSent:     10,
									ResponsesRejected: 1,
								},
							},
						},
						{
							Id:                     "2",
							ControlPlaneInstanceId: "node-002",
							ConnectTime:            util_proto.MustTimestampProto(t2),
							Status: mesh_proto.DiscoverySubscriptionStatus{
								Total: mesh_proto.DiscoveryServiceStats{
									ResponsesSent:     20,
									ResponsesRejected: 2,
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
				Spec: mesh_proto.DataplaneInsight{
					Subscriptions: []*mesh_proto.DiscoverySubscription{
						{
							Id:                     "1",
							ControlPlaneInstanceId: "node-001",
						},
						{
							Id:                     "2",
							ControlPlaneInstanceId: "node-002",
						},
						{
							Id:                     "3",
							ControlPlaneInstanceId: "node-003",
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh:      "pilot",
					Namespace: "default",
					Name:      "simple",
				},
				Spec: mesh_proto.DataplaneInsight{
					Subscriptions: []*mesh_proto.DiscoverySubscription{
						{
							Id:                     "1",
							ControlPlaneInstanceId: "node-001",
						},
					},
				},
			},
		}
	})

	Describe("TablePrinter", func() {

		var dataplaneInsights *mesh_core.DataplaneInsightResourceList
		var buf *bytes.Buffer

		BeforeEach(func() {
			dataplaneInsights = &mesh_core.DataplaneInsightResourceList{}
			buf = &bytes.Buffer{}
		})

		It("should not fail on empty list", func() {
			// given
			dataplaneInsights.Items = nil

			// when
			err := get.PrintDataplaneInsights(now, dataplaneInsights, buf)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(`
MESH   NAME   STATUS   LAST CONNECTED AGO   LAST UPDATED AGO   TOTAL UPDATES   TOTAL ERRORS
`)))
		})

		It("should print a list of 2 items", func() {
			// given
			dataplaneInsights.Items = sampleDataplaneInsights

			// when
			err := get.PrintDataplaneInsights(now, dataplaneInsights, buf)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(`
MESH      NAME         STATUS    LAST CONNECTED AGO   LAST UPDATED AGO   TOTAL UPDATES   TOTAL ERRORS
default   experiment   Online    2h3m4s               never              30              3
default   example      Offline   never                never              0               0
pilot     simple       Offline   never                never              0               0
`)))
		})
	})

	Describe("GetDataplanesCmd", func() {

		var rootCtx *konvoyctl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup

			rootCtx = &konvoyctl_cmd.RootContext{
				Runtime: konvoyctl_cmd.RootRuntime{
					Now: func() time.Time { return now },
					NewResourceStore: func(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, ds := range sampleDataplaneInsights {
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

		DescribeTable("konvoyctl get dataplanes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
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
