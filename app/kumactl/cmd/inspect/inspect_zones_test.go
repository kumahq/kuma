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

	"github.com/kumahq/kuma/v3/app/kumactl/cmd"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/resources"
	test_kumactl "github.com/kumahq/kuma/v3/app/kumactl/pkg/test"
	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	util_http "github.com/kumahq/kuma/v3/pkg/util/http"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

type testZoneOverviewClient struct {
	total     uint32
	overviews []*zone_api.ZoneOverviewResource
}

func (c *testZoneOverviewClient) List(_ context.Context) (*zone_api.ZoneOverviewResourceList, error) {
	return &zone_api.ZoneOverviewResourceList{
		Items: c.overviews,
		Pagination: model.Pagination{
			Total: c.total,
		},
	}, nil
}

var _ resources.ZoneOverviewClient = &testZoneOverviewClient{}

var _ = Describe("kumactl inspect zones", func() {
	var now, t1, t2 time.Time
	var sampleZoneOverview []*zone_api.ZoneOverviewResource

	BeforeEach(func() {
		now, _ = time.ParseInLocation(time.RFC3339, "2019-07-17T18:08:41+00:00", time.UTC)
		t1, _ = time.ParseInLocation(time.RFC3339, "2018-07-17T16:05:36.995+00:00", time.UTC)
		t2, _ = time.ParseInLocation(time.RFC3339, "2019-07-17T16:05:36.995+00:00", time.UTC)

		sampleZoneOverview = []*zone_api.ZoneOverviewResource{
			{
				Meta: &test_model.ResourceMeta{
					Name:             "zone-1",
					CreationTime:     t1,
					ModificationTime: now,
				},
				Spec: &zone_api.ZoneOverview{
					Zone: &zone_api.Zone{Enabled: pointer.To(true)},
					ZoneInsight: &zoneinsight_api.ZoneInsight{
						Subscriptions: []*zoneinsight_api.KDSSubscription{
							{
								ID:               "1",
								GlobalInstanceID: "node-001",
								ConnectTime:      zoneinsight_api.NewTime(t1),
								Status: &zoneinsight_api.KDSSubscriptionStatus{
									Total: &zoneinsight_api.KDSServiceStats{
										ResponsesSent:     22,
										ResponsesRejected: 11,
									},
									Stat: map[string]*zoneinsight_api.KDSServiceStats{
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
										"Secret": {
											ResponsesSent:     2,
											ResponsesRejected: 1,
										},
									},
								},
								Version: &zoneinsight_api.Version{
									KumaCP: &zoneinsight_api.KumaCpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
								},
							},
							{
								ID:               "2",
								GlobalInstanceID: "node-002",
								ConnectTime:      zoneinsight_api.NewTime(t2),
								Status: &zoneinsight_api.KDSSubscriptionStatus{
									Total: &zoneinsight_api.KDSServiceStats{
										ResponsesSent:     20,
										ResponsesRejected: 2,
									},
								},
								Version: &zoneinsight_api.Version{
									KumaCP: &zoneinsight_api.KumaCpVersion{
										Version:   "1.0.0",
										GitTag:    "v1.0.0",
										GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
										BuildDate: "2019-08-07T11:26:06Z",
									},
								},
								Config: `{"apiServer":{"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":15681},"https":{"enabled":false,"interface":"0.0.0.0","port":5682,"tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":15678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":15680},"dnsServer":{"domain":"mesh","port":15653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":15678,"tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key","workDir":"/Users/jakob/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":15676,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"tlsCertFile":"/Users/jakob/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/jakob/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:35685","kds":{"maxMsgSize":10485760,"rootCaFile":""},"name":"cluster-1"}},"reports":{"enabled":false},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]}},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}`,
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
				Spec: &zone_api.ZoneOverview{
					Zone: &zone_api.Zone{Enabled: pointer.To(true)},
					ZoneInsight: &zoneinsight_api.ZoneInsight{
						Subscriptions: []*zoneinsight_api.KDSSubscription{
							{
								ID:               "1",
								GlobalInstanceID: "node-001",
							},
							{
								ID:               "2",
								GlobalInstanceID: "node-002",
							},
							{
								ID:               "3",
								GlobalInstanceID: "node-003",
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
				Spec: &zone_api.ZoneOverview{
					Zone: &zone_api.Zone{Enabled: pointer.To(false)},
					ZoneInsight: &zoneinsight_api.ZoneInsight{
						Subscriptions: []*zoneinsight_api.KDSSubscription{
							{
								ID:               "1",
								GlobalInstanceID: "node-001",
								ConnectTime:      zoneinsight_api.NewTime(t2),
								Version: &zoneinsight_api.Version{
									KumaCP: &zoneinsight_api.KumaCpVersion{
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
					"inspect", "zones",
				}, given.outputFormat))

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
