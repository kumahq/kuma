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

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	system_core "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_http "github.com/kumahq/kuma/pkg/util/http"
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
								Config: `{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":15681},"https":{"enabled":false,"interface":"0.0.0.0","port":5682,"tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":15678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":15680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":15653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":15678,"tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key","workDir":"/Users/jakob/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":15676,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:35685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"cluster-1"}},"reports":{"enabled":false},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}`,
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
			rootCtx.Runtime.NewZoneOverviewClient = func(util_http.Client) resources.ZoneOverviewClient {
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
				Expect(buf.String()).To(matchers.MatchGoldenEqual("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "inspect-zones.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-zones.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-zones.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-zone.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
