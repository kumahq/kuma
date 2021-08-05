package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("HealthCheckConfigurer", func() {

	type testCase struct {
		clusterName string
		healthCheck *mesh_core.HealthCheckResource
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.HealthCheck(mesh_core.ProtocolHTTP, given.healthCheck)).
				Configure(clusters.Timeout(mesh_core.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("HealthCheck with neither active nor passive checks", testCase{
			clusterName: "testCluster",
			healthCheck: mesh_core.NewHealthCheckResource(),
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with active checks", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:           durationpb.New(5 * time.Second),
						Timeout:            durationpb.New(4 * time.Second),
						UnhealthyThreshold: 3,
						HealthyThreshold:   2,
						ReuseConnection:    &wrapperspb.BoolValue{Value: false},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              reuseConnection: false
              tcpHealthCheck: {}
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with provided TCP Send/Receive properties", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "frontend"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:           durationpb.New(5 * time.Second),
						Timeout:            durationpb.New(4 * time.Second),
						UnhealthyThreshold: 3,
						HealthyThreshold:   2,
						Tcp: &mesh_proto.HealthCheck_Conf_Tcp{
							Send: &wrapperspb.BytesValue{
								Value: []byte("foo"),
							},
							Receive: []*wrapperspb.BytesValue{
								{Value: []byte("bar")},
								{Value: []byte("baz")},
							},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck:
                receive:
                - text: "626172"
                - text: 62617a
                send:
                  text: 666f6f
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with provided TCP Send only properties", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "frontend"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:           durationpb.New(5 * time.Second),
						Timeout:            durationpb.New(4 * time.Second),
						UnhealthyThreshold: 3,
						HealthyThreshold:   2,
						Tcp: &mesh_proto.HealthCheck_Conf_Tcp{
							Send: &wrapperspb.BytesValue{
								Value: []byte("foo"),
							},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck:
                send:
                  text: 666f6f
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with provided HTTP configuration", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "frontend"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:           durationpb.New(5 * time.Second),
						Timeout:            durationpb.New(4 * time.Second),
						UnhealthyThreshold: 3,
						HealthyThreshold:   2,
						Http: &mesh_proto.HealthCheck_Conf_Http{
							Path: "/foo",
							RequestHeadersToAdd: []*mesh_proto.
								HealthCheck_Conf_Http_HeaderValueOption{
								{
									Header: &mesh_proto.HealthCheck_Conf_Http_HeaderValue{
										Key:   "foobar",
										Value: "foobaz",
									},
									Append: &wrapperspb.BoolValue{Value: false},
								},
							},
							ExpectedStatuses: []*wrapperspb.UInt32Value{
								{Value: 200},
								{Value: 201},
							},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              httpHealthCheck:
                expectedStatuses:
                - end: "201"
                  start: "200"
                - end: "202"
                  start: "201"
                path: /foo
                requestHeadersToAdd:
                - append: false
                  header:
                    key: foobar
                    value: foobaz
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with provided both, TCP and HTTP configurations", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "frontend"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:           durationpb.New(5 * time.Second),
						Timeout:            durationpb.New(4 * time.Second),
						UnhealthyThreshold: 3,
						HealthyThreshold:   2,
						Http: &mesh_proto.HealthCheck_Conf_Http{
							Path: "/foo",
							RequestHeadersToAdd: []*mesh_proto.
								HealthCheck_Conf_Http_HeaderValueOption{
								{
									Header: &mesh_proto.HealthCheck_Conf_Http_HeaderValue{
										Key:   "foobar",
										Value: "foobaz",
									},
									Append: &wrapperspb.BoolValue{Value: false},
								},
							},
							ExpectedStatuses: []*wrapperspb.UInt32Value{
								{Value: 200},
								{Value: 201},
							},
						},
						Tcp: &mesh_proto.HealthCheck_Conf_Tcp{
							Send: &wrapperspb.BytesValue{
								Value: []byte("foo"),
							},
							Receive: []*wrapperspb.BytesValue{
								{Value: []byte("bar")},
								{Value: []byte("baz")},
							},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck:
                receive:
                - text: "626172"
                - text: 62617a
                send:
                  text: 666f6f
              timeout: 4s
              unhealthyThreshold: 3
            - healthyThreshold: 2
              httpHealthCheck:
                expectedStatuses:
                - end: "201"
                  start: "200"
                - end: "202"
                  start: "201"
                path: /foo
                requestHeadersToAdd:
                - append: false
                  header:
                    key: foobar
                    value: foobaz
              interval: 5s
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with jitter", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:              durationpb.New(5 * time.Second),
						Timeout:               durationpb.New(4 * time.Second),
						UnhealthyThreshold:    3,
						HealthyThreshold:      2,
						InitialJitter:         durationpb.New(6 * time.Second),
						IntervalJitter:        durationpb.New(7 * time.Second),
						IntervalJitterPercent: 50,
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck: {}
              timeout: 4s
              unhealthyThreshold: 3
              initialJitter: 6s
              intervalJitter: 7s
              intervalJitterPercent: 50
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with panic threshold", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:              durationpb.New(5 * time.Second),
						Timeout:               durationpb.New(4 * time.Second),
						UnhealthyThreshold:    3,
						HealthyThreshold:      2,
						HealthyPanicThreshold: &wrapperspb.FloatValue{Value: 90},
					},
				},
			},
			expected: `
            commonLbConfig:
              healthyPanicThreshold:
                value: 90
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck: {}
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with panic threshold, fail traffic on panic", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:              durationpb.New(5 * time.Second),
						Timeout:               durationpb.New(4 * time.Second),
						UnhealthyThreshold:    3,
						HealthyThreshold:      2,
						HealthyPanicThreshold: &wrapperspb.FloatValue{Value: 90},
						FailTrafficOnPanic:    &wrapperspb.BoolValue{Value: true},
					},
				},
			},
			expected: `
            commonLbConfig:
              healthyPanicThreshold:
                value: 90
              zoneAwareLbConfig:
                failTrafficOnPanic: true
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck: {}
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with event log path", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:                     durationpb.New(5 * time.Second),
						Timeout:                      durationpb.New(4 * time.Second),
						NoTrafficInterval:            durationpb.New(6 * time.Second),
						UnhealthyThreshold:           3,
						HealthyThreshold:             2,
						EventLogPath:                 "/event/log/path",
						AlwaysLogHealthCheckFailures: &wrapperspb.BoolValue{Value: true},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            healthChecks:
            - alwaysLogHealthCheckFailures: true
              eventLogPath: /event/log/path
              healthyThreshold: 2
              interval: 5s
              noTrafficInterval: 6s
              tcpHealthCheck: {}
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
	)
})
