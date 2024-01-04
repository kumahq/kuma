package server_test

import (
	"context"
	"sync"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	protov1 "github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/grpc"
	kds_samples "github.com/kumahq/kuma/pkg/test/kds/samples"
	kds_setup "github.com/kumahq/kuma/pkg/test/kds/setup"
	kds_verifier "github.com/kumahq/kuma/pkg/test/kds/verifier"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var node = &envoy_core.Node{
	Id:      "test-id",
	Cluster: "test-cluster",
}

const (
	defaultTimeout = 3 * time.Second
)

var _ = Describe("KDS Server", func() {
	var tc kds_verifier.TestContext
	BeforeEach(func() {
		s := memory.NewStore()

		wg := &sync.WaitGroup{}
		wg.Add(1)
		srv, err := kds_setup.NewKdsServerBuilder(s).AsZone("test-cluster").Sotw()
		Expect(err).ToNot(HaveOccurred())
		stream := grpc.NewMockServerStream()
		go func() {
			defer func() {
				wg.Done()
				GinkgoRecover()
			}()
			Expect(srv.StreamKumaResources(stream)).ToNot(HaveOccurred())
		}()

		tc = &kds_verifier.TestContextImpl{
			ResourceStore:      s,
			MockStream:         stream,
			Wg:                 wg,
			Responses:          map[string]*envoy_sd.DiscoveryResponse{},
			LastACKedResponses: map[string]*envoy_sd.DiscoveryResponse{},
		}
	})

	It("should support all existing resource types", func() {
		ctx := context.Background()
		// Do not forget to update this test after updating protobuf `KumaKdsOptions`.
		Expect(registry.Global().ObjectTypes(model.HasKdsEnabled())).
			To(HaveLen(len([]protov1.Message{
				kds_samples.CircuitBreaker,
				kds_samples.Dataplane,
				kds_samples.DataplaneInsight,
				kds_samples.ServiceInsight,
				kds_samples.ExternalService,
				kds_samples.FaultInjection,
				kds_samples.GlobalSecret,
				kds_samples.HealthCheck,
				kds_samples.Mesh1,
				kds_samples.ProxyTemplate,
				kds_samples.RateLimit,
				kds_samples.Retry,
				kds_samples.Secret,
				kds_samples.Timeout,
				kds_samples.TrafficLog,
				kds_samples.TrafficPermission,
				kds_samples.TrafficRoute,
				kds_samples.TrafficTrace,
				kds_samples.ZoneIngress,
				kds_samples.ZoneIngressInsight,
				kds_samples.ZoneEgress,
				kds_samples.ZoneEgressInsight,
				kds_samples.Config,
				kds_samples.VirtualOutbound,
				kds_samples.Gateway,
				kds_samples.GatewayRoute,
			})))

		vrf := kds_verifier.New().
			// NOTE: The resources all have to be created before any DiscoveryRequests are made.
			// This is because in the initial state, there are no snapshots. The first DiscoveryRequest
			// won't get a response until the first snapshot is generated after the reconciliation
			// period expires. Then all the subsequent DiscoveryRequests will be served from the
			// same snapshot, so resources that aren't already present won't be reported.

			Exec(kds_verifier.Create(ctx, &mesh.CircuitBreakerResource{Spec: kds_samples.CircuitBreaker}, store.CreateByKey("cb-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.DataplaneInsightResource{Spec: kds_samples.DataplaneInsight}, store.CreateByKey("insight-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.DataplaneResource{Spec: kds_samples.Dataplane}, store.CreateByKey("dp-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.ExternalServiceResource{Spec: kds_samples.ExternalService}, store.CreateByKey("es-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: kds_samples.FaultInjection}, store.CreateByKey("fi-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.HealthCheckResource{Spec: kds_samples.HealthCheck}, store.CreateByKey("hc-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh-1", model.NoMesh))).
			Exec(kds_verifier.Create(ctx, &mesh.ProxyTemplateResource{Spec: kds_samples.ProxyTemplate}, store.CreateByKey("pt-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.RateLimitResource{Spec: kds_samples.RateLimit}, store.CreateByKey("rl-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.RetryResource{Spec: kds_samples.Retry}, store.CreateByKey("retry-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TimeoutResource{Spec: kds_samples.Timeout}, store.CreateByKey("timeout-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficLogResource{Spec: kds_samples.TrafficLog}, store.CreateByKey("tl-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficPermissionResource{Spec: kds_samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficRouteResource{Spec: kds_samples.TrafficRoute}, store.CreateByKey("tr-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficTraceResource{Spec: kds_samples.TrafficTrace}, store.CreateByKey("tt-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &system.SecretResource{Spec: kds_samples.Secret}, store.CreateByKey("s-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Mesh1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.DataplaneType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Dataplane))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.DataplaneInsightType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.DataplaneInsight))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.ExternalServiceType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.ExternalService))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.CircuitBreakerType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.CircuitBreaker))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.FaultInjection))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.HealthCheckType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.HealthCheck))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficLogType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.TrafficLog))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficPermissionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.TrafficPermission))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficRouteType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.TrafficRoute))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficTraceType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.TrafficTrace))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.ProxyTemplateType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.ProxyTemplate))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, system.SecretType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Secret))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.RetryType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Retry))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TimeoutType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Timeout))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.RateLimitType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.RateLimit))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})

	It("should accept request independently for each type", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: kds_samples.FaultInjection}, store.CreateByKey("FaultInjection", "mesh-2"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
			})).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
			})).
			Exec(kds_verifier.ACK(node, mesh.MeshType)).
			Exec(kds_verifier.ACK(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})

	It("should send response for resources created after DiscoveryRequest", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
			})).
			Exec(kds_verifier.ACK(node, mesh.MeshType)).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})

	It("should send response for resources created before ACK", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
			})).
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh2}, store.CreateByKey("mesh-2", "mesh-2"))).
			Exec(kds_verifier.ACK(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(2))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})

	It("should support update", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Mesh1))
			})).
			Exec(kds_verifier.ACK(node, mesh.MeshType)).
			Exec(func(tc kds_verifier.TestContext) error {
				meshRes := mesh.NewMeshResource()
				if err := tc.Store().Get(ctx, meshRes, store.GetByKey("mesh-1", "mesh-1")); err != nil {
					return err
				}
				meshRes.Spec = kds_samples.Mesh2
				return tc.Store().Update(ctx, meshRes)
			}).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(MatchProto(kds_samples.Mesh2))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})

	It("should have deterministic MarshalAny to avoid excess snapshot versions", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: kds_samples.FaultInjection}, store.CreateByKey("fi-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
			})).
			Exec(kds_verifier.ACK(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.ExpectNoResponseDuring(200 * time.Millisecond)).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})

	It("should repeat DiscoveryResponse after NACK", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
			})).
			Exec(kds_verifier.ACK(node, mesh.MeshType)).
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh2}, store.CreateByKey("mesh2", "mesh-2"))).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(2))
			})).
			Exec(kds_verifier.NACK(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(2))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		tc.WaitGroup().Wait()
	})
})
