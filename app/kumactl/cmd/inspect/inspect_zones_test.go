package inspect_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	system_core "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type testZoneOverviewClient struct {
	total     uint32
	overviews []*system_core.ZoneOverviewResource
}

func (c *testZoneOverviewClient) List(_ context.Context) (*system_core.ZoneOverviewResourceList, error) {
	return &system_core.ZoneOverviewResourceList{
		Items: c.overviews,
		Pagination: model.Pagination{
			Total: c.total,
		},
	}, nil
}

var _ resources.ZoneOverviewClient = &testZoneOverviewClient{}

var _ = Describe("kumactl inspect zones", func() {

	var now, t1, t2 time.Time
	var sampleZoneOverview []*system_core.ZoneOverviewResource

	BeforeEach(func() {
		now, _ = time.Parse(time.RFC3339, "2019-07-17T18:08:41+00:00")
		t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		t2, _ = time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")
		time.Local = time.UTC

		sampleZoneOverview = []*system_core.ZoneOverviewResource{
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-1",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &system_proto.ZoneOverview{
					Zone: &system_proto.Zone{Enabled: util_proto.Bool(true)},
					ZoneInsight: &system_proto.ZoneInsight{
						Subscriptions: []*system_proto.KDSSubscription{
							{
								Id:               "1",
								GlobalInstanceId: "node-001",
								ConnectTime:      util_proto.MustTimestampProto(t1),
								Status: &system_proto.KDSSubscriptionStatus{
									Total: &system_proto.KDSServiceStats{
										ResponsesSent:     22,
										ResponsesRejected: 11,
									},
									Stat: map[string]*system_proto.KDSServiceStats{
										"Mesh": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"Ingress": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"FaultInjection": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"CircuitBreaker": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"HealthCheck": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"RateLimit": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"TrafficTrace": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"TrafficRoute": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"TrafficPermission": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"TrafficLog": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"ProxyTemplate": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
										"Secret": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
									},
								},
								Version: &system_proto.Version{
									KumaCp: &system_proto.KumaCpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
								},
							},
							{
								Id:               "2",
								GlobalInstanceId: "node-002",
								ConnectTime:      util_proto.MustTimestampProto(t2),
								Status: &system_proto.KDSSubscriptionStatus{
									Total: &system_proto.KDSServiceStats{
										ResponsesSent:     20,
										ResponsesRejected: 2,
									},
								},
								Version: &system_proto.Version{
									KumaCp: &system_proto.KumaCpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
								},
							},
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-2",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &system_proto.ZoneOverview{
					Zone: &system_proto.Zone{Enabled: util_proto.Bool(true)},
					ZoneInsight: &system_proto.ZoneInsight{
						Subscriptions: []*system_proto.KDSSubscription{
							{
								Id:               "1",
								GlobalInstanceId: "node-001",
							},
							{
								Id:               "2",
								GlobalInstanceId: "node-002",
							},
							{
								Id:               "3",
								GlobalInstanceId: "node-003",
							},
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-3",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &system_proto.ZoneOverview{
					Zone: &system_proto.Zone{Enabled: util_proto.Bool(false)},
					ZoneInsight: &system_proto.ZoneInsight{
						Subscriptions: []*system_proto.KDSSubscription{
							{
								Id:               "1",
								GlobalInstanceId: "node-001",
								ConnectTime:      util_proto.MustTimestampProto(t2),
								Version: &system_proto.Version{
									KumaCp: &system_proto.KumaCpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
								},
							},
						},
					},
				},
			},
		}
	})

	Describe("InspectZonesCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer

		var testClient *testZoneOverviewClient

		BeforeEach(func() {
			// setup
			testClient = &testZoneOverviewClient{
				total:     uint32(len(sampleZoneOverview)),
				overviews: sampleZoneOverview,
			}
			rootCtx, err := test_kumactl.MakeRootContext(now, nil)
			Expect(err).ToNot(HaveOccurred())
			rootCtx.Runtime.NewZoneOverviewClient = func(server *config_proto.ControlPlaneCoordinates_ApiServer) (resources.ZoneOverviewClient, error) {
				return testClient, nil
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

		DescribeTable("kumactl inspect zones -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"inspect", "zones"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.String()).To(matchers.MatchGoldenEqual(filepath.Join("testdata", given.goldenFile)))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "inspect-zones.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-zones.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-zones.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-zone.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
