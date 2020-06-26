package server_test

import (
	"context"
	"sync"
	"time"

	"github.com/Kong/kuma/pkg/core/resources/model"
	kds_samples "github.com/Kong/kuma/pkg/test/kds/samples"
	kds_setup "github.com/Kong/kuma/pkg/test/kds/setup"

	"github.com/gogo/protobuf/proto"

	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/kds"
	kds_verifier "github.com/Kong/kuma/pkg/test/kds/verifier"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var (
	node = &envoy_core.Node{
		Id:      "test-id",
		Cluster: "test-cluster",
	}
)

const (
	defaultTimeout = 3 * time.Second
)

var _ = Describe("KDS Server", func() {

	var tc kds_verifier.TestContext
	BeforeEach(func() {
		s := memory.NewStore()

		wg := &sync.WaitGroup{}
		wg.Add(1)
		stream := kds_setup.StartServer(s, wg)

		tc = &kds_verifier.TestContextImpl{
			ResourceStore:      s,
			MockStream:         stream,
			Wg:                 wg,
			Responses:          map[string]*v2.DiscoveryResponse{},
			LastACKedResponses: map[string]*v2.DiscoveryResponse{},
		}
	})

	It("should support all existing resource types", func() {
		ctx := context.Background()

		// Just to don't forget to update this test after updating 'kds.SupportedTypes
		Expect([]proto.Message{&kds_samples.Mesh1, &kds_samples.Ingress, &kds_samples.CircuitBreaker, &kds_samples.FaultInjection, &kds_samples.HealthCheck, &kds_samples.TrafficLog, &kds_samples.TrafficPermission, &kds_samples.TrafficRoute, &kds_samples.TrafficTrace, &kds_samples.ProxyTemplate, &kds_samples.Secret}).To(HaveLen(len(kds.SupportedTypes)))

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: kds_samples.Mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.DataplaneResource{Spec: kds_samples.Ingress}, store.CreateByKey("Ingress-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.CircuitBreakerResource{Spec: kds_samples.CircuitBreaker}, store.CreateByKey("cb-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: kds_samples.FaultInjection}, store.CreateByKey("fi-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.HealthCheckResource{Spec: kds_samples.HealthCheck}, store.CreateByKey("hc-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficLogResource{Spec: kds_samples.TrafficLog}, store.CreateByKey("tl-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficPermissionResource{Spec: kds_samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficRouteResource{Spec: kds_samples.TrafficRoute}, store.CreateByKey("tr-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficTraceResource{Spec: kds_samples.TrafficTrace}, store.CreateByKey("tt-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.ProxyTemplateResource{Spec: kds_samples.ProxyTemplate}, store.CreateByKey("pt-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &system.SecretResource{Spec: kds_samples.Secret}, store.CreateByKey("s-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.Mesh1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.DataplaneType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.Ingress))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.CircuitBreakerType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.CircuitBreaker))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.FaultInjection))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.HealthCheckType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.HealthCheck))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficLogType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.TrafficLog))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficPermissionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.TrafficPermission))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficRouteType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.TrafficRoute))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficTraceType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.TrafficTrace))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.ProxyTemplateType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.ProxyTemplate))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, system.SecretType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.Secret))
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
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.Mesh1))
			})).
			Exec(kds_verifier.ACK(node, mesh.MeshType)).
			Exec(func(tc kds_verifier.TestContext) error {
				var meshRes mesh.MeshResource
				if err := tc.Store().Get(ctx, &meshRes, store.GetByKey("mesh-1", "mesh-1")); err != nil {
					return err
				}
				meshRes.Spec = kds_samples.Mesh2
				if err := tc.Store().Update(ctx, &meshRes); err != nil {
					return err
				}
				return nil
			}).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(rs []model.Resource) {
				Expect(rs).To(HaveLen(1))
				Expect(rs[0].GetSpec()).To(Equal(&kds_samples.Mesh2))
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
