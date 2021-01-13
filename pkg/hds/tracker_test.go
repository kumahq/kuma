package hds_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_service_health_v3 "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/hds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HDS Tracker", func() {

	var tracker hds.Callbacks
	var resourceManager manager.ResourceManager

	createDp := func() {
		err := resourceManager.Create(context.Background(), &mesh.DataplaneResource{Spec: &proto.Dataplane{
			Networking: &proto.Dataplane_Networking{
				Address: "10.20.0.1",
				Inbound: []*proto.Dataplane_Networking_Inbound{{
					Port:           9000,
					ServiceAddress: "192.168.0.1",
					ServicePort:    80,
					Tags: map[string]string{
						proto.ServiceTag: "backend",
					}},
				},
			},
		}}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {
		resourceManager = manager.NewResourceManager(memory.NewStore())
		tracker = hds.NewTracker(resourceManager)

		err := resourceManager.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return HealthCheckSpecifier", func() {
		createDp()

		hcs, err := tracker.OnHealthCheckRequest(1,
			&envoy_service_health_v3.HealthCheckRequest{
				Node: &envoy_core.Node{Id: "mesh-1.dp-1"},
			})
		Expect(err).ToNot(HaveOccurred())

		expected := `
clusterHealthChecks:
- clusterName: localhost:80
  healthChecks:
  - healthyThreshold: 1
    interval: 1s
    noTrafficInterval: 1s
    tcpHealthCheck: {}
    timeout: 2s
    unhealthyThreshold: 1
  localityEndpoints:
  - endpoints:
    - address:
        socketAddress:
          address: 192.168.0.1
          portValue: 80`

		actual, err := util_proto.ToYAML(hcs)
		Expect(actual).To(MatchYAML(expected))
	})

	It("should change Dataplane health", func() {
		createDp()

		_, err := tracker.OnHealthCheckRequest(1,
			&envoy_service_health_v3.HealthCheckRequest{
				Node: &envoy_core.Node{Id: "mesh-1.dp-1"},
			})
		Expect(err).ToNot(HaveOccurred())

		err = tracker.OnEndpointHealthResponse(1,
			&envoy_service_health_v3.EndpointHealthResponse{
				ClusterEndpointsHealth: []*envoy_service_health_v3.ClusterEndpointsHealth{{
					ClusterName: "localhost:80",
					LocalityEndpointsHealth: []*envoy_service_health_v3.LocalityEndpointsHealth{{
						EndpointsHealth: []*envoy_service_health_v3.EndpointHealth{{
							Endpoint: &envoy_endpoint.Endpoint{
								Address: &envoy_core.Address{
									Address: &envoy_core.Address_SocketAddress{
										SocketAddress: &envoy_core.SocketAddress{
											PortSpecifier: &envoy_core.SocketAddress_PortValue{PortValue: 80},
										},
									},
								},
							},
							HealthStatus: envoy_core.HealthStatus_UNHEALTHY,
						}},
					}},
				}},
			})
		Expect(err).ToNot(HaveOccurred())

		dp := mesh.NewDataplaneResource()
		err = resourceManager.Get(context.Background(), dp, store.GetByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
		Expect(dp.Spec.Networking.Inbound[0].Health).ToNot(BeNil())
		Expect(dp.Spec.Networking.Inbound[0].Health.Ready).To(BeFalse())
	})
})
