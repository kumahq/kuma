package inspect_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type testZoneIngressOverviewClient struct {
	total     uint32
	overviews []*core_mesh.ZoneIngressOverviewResource
}

func (c *testZoneIngressOverviewClient) List(_ context.Context) (*core_mesh.ZoneIngressOverviewResourceList, error) {
	return &core_mesh.ZoneIngressOverviewResourceList{
		Items: c.overviews,
		Pagination: model.Pagination{
			Total: c.total,
		},
	}, nil
}

var _ resources.ZoneIngressOverviewClient = &testZoneIngressOverviewClient{}

var _ = Describe("kumactl inspect zone-ingresses", func() {

	var now, t1, t2 time.Time
	var sampleZoneIngressOverview []*core_mesh.ZoneIngressOverviewResource

	BeforeEach(func() {
		now, _ = time.Parse(time.RFC3339, "2019-07-17T18:08:41+00:00")
		t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		t2, _ = time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")
		time.Local = time.UTC

		sampleZoneIngressOverview = []*core_mesh.ZoneIngressOverviewResource{
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-ingress-1",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &mesh_proto.ZoneIngressOverview{
					ZoneIngress: &mesh_proto.ZoneIngress{},
					ZoneIngressInsight: &mesh_proto.ZoneIngressInsight{
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
								Version: &mesh_proto.Version{
									KumaDp: &mesh_proto.KumaDpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
									Envoy: &mesh_proto.EnvoyVersion{
										Version: "1.16.0",
										Build:   "hash/1.16.0/RELEASE",
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
								Version: &mesh_proto.Version{
									KumaDp: &mesh_proto.KumaDpVersion{
										Version:   "1.0.2",
										GitTag:    "v1.0.2",
										GitCommit: "9d868cd8681c4021bb4a10bf2306ca613ba4de42",
										BuildDate: "2020-08-07T11:26:06Z",
									},
									Envoy: &mesh_proto.EnvoyVersion{
										Version: "1.16.1",
										Build:   "hash/1.16.1/RELEASE",
									},
								},
							},
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-ingress-2",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &mesh_proto.ZoneIngressOverview{
					ZoneIngress: &mesh_proto.ZoneIngress{},
					ZoneIngressInsight: &mesh_proto.ZoneIngressInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								Id:                     "1",
								ControlPlaneInstanceId: "node-001",
								ConnectTime:            util_proto.MustTimestampProto(t1),
								DisconnectTime:         util_proto.MustTimestampProto(t1.Add(1 * time.Minute)),
								Status: &mesh_proto.DiscoverySubscriptionStatus{
									Total: &mesh_proto.DiscoveryServiceStats{
										ResponsesSent:     10,
										ResponsesRejected: 1,
									},
								},
								Version: &mesh_proto.Version{
									KumaDp: &mesh_proto.KumaDpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
									Envoy: &mesh_proto.EnvoyVersion{
										Version: "1.16.0",
										Build:   "hash/1.16.0/RELEASE",
									},
								},
							},
							{
								Id:                     "2",
								ControlPlaneInstanceId: "node-002",
								ConnectTime:            util_proto.MustTimestampProto(t2),
								DisconnectTime:         util_proto.MustTimestampProto(t2.Add(1 * time.Minute)),
								Status: &mesh_proto.DiscoverySubscriptionStatus{
									Total: &mesh_proto.DiscoveryServiceStats{
										ResponsesSent:     20,
										ResponsesRejected: 2,
									},
								},
								Version: &mesh_proto.Version{
									KumaDp: &mesh_proto.KumaDpVersion{
										Version:   "1.0.2",
										GitTag:    "v1.0.2",
										GitCommit: "9d868cd8681c4021bb4a10bf2306ca613ba4de42",
										BuildDate: "2020-08-07T11:26:06Z",
									},
									Envoy: &mesh_proto.EnvoyVersion{
										Version: "1.16.1",
										Build:   "hash/1.16.1/RELEASE",
									},
								},
							},
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-ingress-3",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &mesh_proto.ZoneIngressOverview{
					ZoneIngress: &mesh_proto.ZoneIngress{},
					ZoneIngressInsight: &mesh_proto.ZoneIngressInsight{
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
								Version: &mesh_proto.Version{
									KumaDp: &mesh_proto.KumaDpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
									Envoy: &mesh_proto.EnvoyVersion{
										Version: "1.16.0",
										Build:   "hash/1.16.0/RELEASE",
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
								Version: &mesh_proto.Version{
									KumaDp: &mesh_proto.KumaDpVersion{
										Version:   "1.0.2",
										GitTag:    "v1.0.2",
										GitCommit: "9d868cd8681c4021bb4a10bf2306ca613ba4de42",
										BuildDate: "2020-08-07T11:26:06Z",
									},
									Envoy: &mesh_proto.EnvoyVersion{
										Version: "1.16.1",
										Build:   "hash/1.16.1/RELEASE",
									},
								},
							},
						},
					},
				},
			},
		}
	})

	Describe("InspectZoneIngressesCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer

		var testClient *testZoneIngressOverviewClient

		BeforeEach(func() {
			// setup
			testClient = &testZoneIngressOverviewClient{
				total:     uint32(len(sampleZoneIngressOverview)),
				overviews: sampleZoneIngressOverview,
			}

			rootCtx, err := test_kumactl.MakeRootContext(now, nil)
			Expect(err).ToNot(HaveOccurred())
			rootCtx.Runtime.NewZoneIngressOverviewClient = func(util_http.Client) resources.ZoneIngressOverviewClient {
				return testClient
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			matcher      func(path ...string) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl inspect zone-ingresses -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"inspect", "zone-ingresses"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.String()).To(matchers.MatchGoldenEqual("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "inspect-zone-ingresses.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-zone-ingresses.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-zone-ingresses.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-zone-ingress.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
