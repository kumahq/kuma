package inspect_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kong/kuma/app/kumactl/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

type testDataplaneOverviewClient struct {
	receivedTags map[string]string
	overviews    []*mesh_core.DataplaneOverviewResource
}

func (c *testDataplaneOverviewClient) List(_ context.Context, _ string, tags map[string]string) (*mesh_core.DataplaneOverviewResourceList, error) {
	c.receivedTags = tags
	return &mesh_core.DataplaneOverviewResourceList{
		Items: c.overviews,
	}, nil
}

var _ resources.DataplaneOverviewClient = &testDataplaneOverviewClient{}

var _ = Describe("kumactl inspect dataplanes", func() {

	var now, t1, t2 time.Time
	var sampleDataplaneOverview []*mesh_core.DataplaneOverviewResource

	BeforeEach(func() {
		now, _ = time.Parse(time.RFC3339, "2019-07-17T18:08:41+00:00")
		t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		t2, _ = time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")

		sampleDataplaneOverview = []*mesh_core.DataplaneOverviewResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "experiment",
				},
				Spec: mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
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
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								Id:                     "1",
								ControlPlaneInstanceId: "node-001",
								ConnectTime:            util_proto.MustTimestampProto(t1),
								Status: &mesh_proto.DiscoverySubscriptionStatus{
									Total: &mesh_proto.DiscoveryServiceStats{
										ResponsesSent:     10,
										ResponsesRejected: 1,
									},
								},
							},
							{
								Id:                     "2",
								ControlPlaneInstanceId: "node-002",
								ConnectTime:            util_proto.MustTimestampProto(t2),
								Status: &mesh_proto.DiscoverySubscriptionStatus{
									Total: &mesh_proto.DiscoveryServiceStats{
										ResponsesSent:     20,
										ResponsesRejected: 2,
									},
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
				Spec: mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Interface: "127.0.0.1:8080:80",
									Tags: map[string]string{
										"service": "example",
									},
								},
							},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
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
			},
		}
	})

	Describe("GetDataplanesCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer

		var testClient *testDataplaneOverviewClient

		BeforeEach(func() {
			// setup
			testClient = &testDataplaneOverviewClient{
				overviews: sampleDataplaneOverview,
			}

			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: func() time.Time { return now },
					NewDataplaneOverviewClient: func(*config_proto.ControlPlaneCoordinates_ApiServer) (resources.DataplaneOverviewClient, error) {
						return testClient, nil
					},
				},
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

		DescribeTable("kumactl inspect dataplanes -o table|json|yaml",
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
				goldenFile:   "inspect-dataplanes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-dataplanes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-dataplanes.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-dataplanes.golden.yaml",
				matcher:      MatchYAML,
			}),
		)

		Describe("kumactl inspect dataplanes --tag", func() {
			It("tags should be passed to the client", func() {
				// given
				rootCmd.SetArgs([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "dataplanes", "--tag", "service=mobile", "--tag", "version=v1"})

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(testClient.receivedTags).To(HaveKeyWithValue("service", "mobile"))
				Expect(testClient.receivedTags).To(HaveKeyWithValue("version", "v1"))
			})
		})
	})
})
