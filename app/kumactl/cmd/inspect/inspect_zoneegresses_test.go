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

type testZoneEgressOverviewClient struct {
	total     uint32
	overviews []*core_mesh.ZoneEgressOverviewResource
}

func (c *testZoneEgressOverviewClient) List(_ context.Context) (*core_mesh.ZoneEgressOverviewResourceList, error) {
	return &core_mesh.ZoneEgressOverviewResourceList{
		Items: c.overviews,
		Pagination: model.Pagination{
			Total: c.total,
		},
	}, nil
}

var _ resources.ZoneEgressOverviewClient = &testZoneEgressOverviewClient{}

var _ = Describe("kumactl inspect zoneegresses", func() {

	var now, t1, t2 time.Time
	var sampleZoneEgressOverview []*core_mesh.ZoneEgressOverviewResource

	BeforeEach(func() {
		now, _ = time.Parse(time.RFC3339, "2019-07-17T18:08:41+00:00")
		t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		t2, _ = time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")
		time.Local = time.UTC

		sampleZoneEgressOverview = []*core_mesh.ZoneEgressOverviewResource{
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zoneegress-1",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &mesh_proto.ZoneEgressOverview{
					ZoneEgress: &mesh_proto.ZoneEgress{},
					ZoneEgressInsight: &mesh_proto.ZoneEgressInsight{
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
					Name:             "zoneegress-2",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &mesh_proto.ZoneEgressOverview{
					ZoneEgress: &mesh_proto.ZoneEgress{},
					ZoneEgressInsight: &mesh_proto.ZoneEgressInsight{
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
					Name:             "zoneegress-3",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &mesh_proto.ZoneEgressOverview{
					ZoneEgress: &mesh_proto.ZoneEgress{},
					ZoneEgressInsight: &mesh_proto.ZoneEgressInsight{
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

	Describe("InspectZoneEgressesCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer

		var testClient *testZoneEgressOverviewClient

		BeforeEach(func() {
			// setup
			testClient = &testZoneEgressOverviewClient{
				total:     uint32(len(sampleZoneEgressOverview)),
				overviews: sampleZoneEgressOverview,
			}

			rootCtx, err := test_kumactl.MakeRootContext(now, nil)
			Expect(err).ToNot(HaveOccurred())
			rootCtx.Runtime.NewZoneEgressOverviewClient = func(util_http.Client) resources.ZoneEgressOverviewClient {
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

		DescribeTable("kumactl inspect zoneegresses -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"inspect", "zoneegresses"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.String()).To(matchers.MatchGoldenEqual("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "inspect-zoneegresses.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-zoneegresses.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-zoneegresses.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-zoneegress.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
