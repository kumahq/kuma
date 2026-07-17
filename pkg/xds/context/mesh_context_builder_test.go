package context_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	test_store "github.com/kumahq/kuma/v3/pkg/test/store"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_server "github.com/kumahq/kuma/v3/pkg/xds/server"
)

var _ = Describe("hash", func() {
	lookupIPFunc := func(s string) ([]net.IP, error) {
		return []net.IP{net.ParseIP(s)}, nil
	}
	var resourceStore store.ResourceStore
	var meshContextBuilder xds_context.MeshContextBuilder

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			nil,
		)
	})

	_ = DescribeTable("with BaseMeshContext", func(inputFile string) {
		// Takes input.yaml compute the hash and then apply a set of changes, then check whether or not it stays the same
		// Given

		inputs, err := os.ReadFile(inputFile)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(inputs), "\n", 2)
		Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
		actions := strings.Split(strings.Trim(parts[0], "# "), " ")
		Expect(actions).To(HaveLen(2), "the first line of the input should be: # <mesh> <bool to indicate if there a change or not>")
		Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

		meshName := strings.TrimSpace(actions[0])
		shouldChange, err := strconv.ParseBool(actions[1])
		Expect(err).ToNot(HaveOccurred())
		beforeContext, err := meshContextBuilder.BuildBaseMeshContextIfChanged(context.Background(), meshName, nil)
		Expect(err).ToNot(HaveOccurred())

		// When
		Expect(test_store.LoadResourcesFromFile(context.Background(), resourceStore, strings.Replace(inputFile, "input", "change", 1))).To(Succeed())

		// Then
		afterContext, err := meshContextBuilder.BuildBaseMeshContextIfChanged(context.Background(), meshName, beforeContext)
		Expect(err).ToNot(HaveOccurred())
		if shouldChange {
			Expect(afterContext.Hash()).ToNot(Equal(beforeContext.Hash()), "context didn't change when it should have")
			Expect(afterContext).ToNot(Equal(beforeContext))
		} else {
			Expect(afterContext.Hash()).To(Equal(beforeContext.Hash()), "context changed when it shouldn't have")
			Expect(afterContext).To(Equal(beforeContext), "context should be the exact same object")
		}
	}, test.EntriesForFolder("basemeshcontext_hash"))

	_ = DescribeTable("with GlobalContext", func(inputFile string) {
		// Takes input.yaml compute the hash and then apply a set of changes, then check whether or not it stays the same
		// Given

		inputs, err := os.ReadFile(inputFile)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(inputs), "\n", 2)
		Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
		actions := strings.Split(strings.Trim(parts[0], "# "), " ")
		Expect(actions).To(HaveLen(1), "the first line of the input should be: # <bool to indicate if there a change or not>")
		Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

		shouldChange, err := strconv.ParseBool(actions[0])
		Expect(err).ToNot(HaveOccurred())

		beforeContext, err := meshContextBuilder.BuildGlobalContextIfChanged(context.Background(), nil)
		Expect(err).ToNot(HaveOccurred())

		// When
		Expect(test_store.LoadResourcesFromFile(context.Background(), resourceStore, strings.Replace(inputFile, "input", "change", 1))).To(Succeed())

		// Then
		afterContext, err := meshContextBuilder.BuildGlobalContextIfChanged(context.Background(), beforeContext)
		Expect(err).ToNot(HaveOccurred())
		if shouldChange {
			Expect(afterContext.Hash()).ToNot(Equal(beforeContext.Hash()), "context didn't change when it should have")
			Expect(afterContext).ToNot(Equal(beforeContext))
		} else {
			Expect(afterContext.Hash()).To(Equal(beforeContext.Hash()), "context changed when it shouldn't have")
			Expect(afterContext).To(Equal(beforeContext), "context should be the exact same object")
		}
	}, test.EntriesForFolder("globalcontext_hash"))

	_ = DescribeTable("with MeshContext", func(inputFile string) {
		// Takes input.yaml compute the hash and then apply a set of changes, then check whether or not it stays the same
		// Given

		inputs, err := os.ReadFile(inputFile)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(inputs), "\n", 2)
		Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
		actions := strings.Split(strings.Trim(parts[0], "# "), " ")
		Expect(actions).To(HaveLen(2), "the first line of the input should be: # <mesh> <bool to indicate if there a change or not>")
		Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

		meshName := strings.TrimSpace(actions[0])
		shouldChange, err := strconv.ParseBool(actions[1])
		Expect(err).ToNot(HaveOccurred())
		beforeContext, err := meshContextBuilder.BuildIfChanged(context.Background(), meshName, nil)
		Expect(err).ToNot(HaveOccurred())

		// When
		Expect(test_store.LoadResourcesFromFile(context.Background(), resourceStore, strings.Replace(inputFile, "input", "change", 1))).To(Succeed())

		// Then
		afterContext, err := meshContextBuilder.BuildIfChanged(context.Background(), meshName, beforeContext)
		Expect(err).ToNot(HaveOccurred())
		if shouldChange {
			Expect(afterContext.Hash).ToNot(Equal(beforeContext.Hash), "context didn't change when it should have")
			Expect(afterContext).ToNot(Equal(beforeContext))
		} else {
			Expect(afterContext.Hash).To(Equal(beforeContext.Hash), "context changed when it shouldn't have")
			Expect(afterContext).To(Equal(beforeContext), "context should be the exact same object")
		}
	}, test.EntriesForFolder("meshcontext_hash"))

	It("should not recompute the mesh context when a Dataplane write only bumps resourceVersion", func() {
		// given a mesh with a single Dataplane
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		before, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		// when the Dataplane is written again with an identical spec (only resourceVersion changes)
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		after, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", before)
		Expect(err).ToNot(HaveOccurred())

		// then the cached context is reused and the mesh hash is unchanged
		Expect(after).To(BeIdenticalTo(before), "resourceVersion-only write should not trigger mesh-wide xDS recomputation")
		Expect(after.Hash).To(Equal(before.Hash))
	})

	It("should not recompute the mesh context when a MeshService write only bumps DataplaneProxies stats", func() {
		// given a mesh with a single MeshService
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: MeshService
name: redis
mesh: mesh-1
spec:
  selector:
    dataplaneTags:
      app: redis
  ports:
  - port: 6739
    appProtocol: tcp
status:
  vips:
  - ip: 10.0.1.1
`)).To(Succeed())

		before, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		// when the MeshService is written again with the same spec/status but updated proxy stats
		meshService := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), meshService, store.GetByKey("redis", "mesh-1"))).To(Succeed())
		meshService.Status.DataplaneProxies = meshservice_api.DataplaneProxies{
			Connected: 3,
			Healthy:   2,
			Total:     3,
		}
		Expect(resourceStore.Update(context.Background(), meshService, store.UpdateWithLabels(meshService.GetMeta().GetLabels()))).To(Succeed())

		after, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", before)
		Expect(err).ToNot(HaveOccurred())

		// then the cached context is reused and the mesh hash is unchanged
		Expect(after).To(BeIdenticalTo(before), "DataplaneProxies-only write should not trigger mesh-wide xDS recomputation")
		Expect(after.Hash).To(Equal(before.Hash))
	})

	It("keeps PolicyMatchingHash stable across resourceVersion-only Dataplane writes", func() {
		builderWithPolicyMatchingHash := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			nil,
			xds_context.WithPolicyMatchingHash(),
		)

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		before, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		after, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(after.PolicyMatchingHash).To(Equal(before.PolicyMatchingHash))
	})

	It("keeps PolicyMatchingHash stable across MeshService DataplaneProxies updates", func() {
		builderWithPolicyMatchingHash := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			nil,
			xds_context.WithPolicyMatchingHash(),
		)

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: MeshService
name: redis
mesh: mesh-1
spec:
  selector:
    dataplaneTags:
      app: redis
  ports:
  - port: 6739
    appProtocol: tcp
status:
  vips:
  - ip: 10.0.1.1
`)).To(Succeed())

		before, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		meshService := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), meshService, store.GetByKey("redis", "mesh-1"))).To(Succeed())
		meshService.Status.DataplaneProxies = meshservice_api.DataplaneProxies{
			Connected: 3,
			Healthy:   2,
			Total:     3,
		}
		Expect(resourceStore.Update(context.Background(), meshService, store.UpdateWithLabels(meshService.GetMeta().GetLabels()))).To(Succeed())

		after, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(after.PolicyMatchingHash).To(Equal(before.PolicyMatchingHash))
	})

	It("recomputes the mesh context when a remote MeshService and its ZoneIngress newly appear", func() {
		// given an mTLS-enabled mesh with MeshServices everywhere, matching the e2e repro
		Expect(samples.MeshMTLSBuilder().
			WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).
			Create(resourceStore)).To(Succeed())

		before, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(before.EndpointMap).To(BeEmpty())

		// when a remote zone's ZoneIngress (with a resolved public address and
		// AvailableServices) and its auto-generated MeshService arrive together,
		// as they would over KDS when a new zone joins
		Expect(builders.ZoneIngress().
			WithZone("east").
			WithAddress("192.168.0.1").
			WithAdvertisedAddress("192.168.0.1").
			WithPort(10001).
			WithAdvertisedPort(10001).
			AddSimpleAvailableService("backend").
			Create(resourceStore)).To(Succeed())
		Expect(samples.MeshServiceSyncedBackendBuilder().Create(resourceStore)).To(Succeed())

		after, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", before)
		Expect(err).ToNot(HaveOccurred())

		// then the mesh context must be rebuilt, with the new remote endpoint present
		Expect(after.Hash).ToNot(Equal(before.Hash), "a newly synced remote MeshService+ZoneIngress must invalidate the mesh context")
		Expect(after).ToNot(BeIdenticalTo(before))
		Expect(after.EndpointMap).ToNot(BeEmpty(), "the remote MeshService should get an endpoint via the ZoneIngress")
	})

	It("recomputes the mesh context through the staged arrival matching the real KDS sequence", func() {
		// given an mTLS-enabled mesh with MeshServices everywhere
		Expect(samples.MeshMTLSBuilder().
			WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).
			Create(resourceStore)).To(Succeed())

		ctx0, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())

		// stage 1: the remote zone's ZoneIngress arrives first, fully resolved
		// (matches the real KDS trace: ZoneIngress observed complete from creationTime)
		Expect(builders.ZoneIngress().
			WithZone("east").
			WithAddress("192.168.0.1").
			WithAdvertisedAddress("192.168.0.1").
			WithPort(10001).
			WithAdvertisedPort(10001).
			Create(resourceStore)).To(Succeed())

		ctx1, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", ctx0)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx1.Hash).ToNot(Equal(ctx0.Hash), "ZoneIngress arrival alone must invalidate the mesh context")
		Expect(ctx1.EndpointMap).To(BeEmpty(), "no MeshService exists yet, so there's nothing to route to")

		// stage 2: the auto-generated MeshService is created next, WITHOUT a VIP yet
		// (matches the real KDS trace: MeshService created ~2s before "vips.allocator: allocating IP")
		Expect(samples.MeshServiceSyncedBackendBuilder().WithoutVIP().Create(resourceStore)).To(Succeed())

		ctx2, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", ctx1)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx2.Hash).ToNot(Equal(ctx1.Hash), "MeshService arrival must invalidate the mesh context")
		Expect(ctx2.EndpointMap).ToNot(BeEmpty(), "cross-zone routing goes through the ZoneIngress and does not require a VIP")

		// stage 3: the VIP is allocated a couple seconds later
		meshService := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), meshService, store.GetByKey(samples.MeshServiceSyncedBackendBuilder().Build().GetMeta().GetName(), "default"))).To(Succeed())
		meshService.Status.VIPs = []meshservice_api.VIP{{IP: "240.0.0.3"}}
		Expect(resourceStore.Update(context.Background(), meshService, store.UpdateWithLabels(meshService.GetMeta().GetLabels()))).To(Succeed())

		ctx3, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", ctx2)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx3.Hash).ToNot(Equal(ctx2.Hash), "VIP allocation must invalidate the mesh context")
	})
})

var _ = Describe("ServicesInformation", func() {
	lookupIPFunc := func(s string) ([]net.IP, error) {
		return []net.IP{net.ParseIP(s)}, nil
	}
	var resourceStore store.ResourceStore
	var meshContextBuilder xds_context.MeshContextBuilder

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			nil,
		)
	})

	It("resolves TLS readiness off MeshService status and treats external services as always ready", func() {
		// given a mesh with a PERMISSIVE mTLS backend
		meshBuilder := builders.Mesh().
			WithBuiltinMTLSBackend("ca-1").
			WithEnabledMTLSBackend("ca-1").
			WithPermissiveMTLSBackends()
		Expect(meshBuilder.Create(resourceStore)).To(Succeed())
		meshName := meshBuilder.Build().GetMeta().GetName()

		// and a MeshService-backed service whose status is TLS ready
		msBuilder := builders.MeshService().
			WithMesh(meshName).
			WithName("backend").
			WithDataplaneTagsSelectorKV(mesh_proto.ServiceTag, "backend").
			AddIntPort(80, 8080, core_meta.ProtocolHTTP).
			WithTLSStatus(meshservice_api.TLSReady)
		Expect(msBuilder.Create(resourceStore)).To(Succeed())
		ms := msBuilder.Build()

		// and a legacy external service, which is always considered TLS ready
		externalService := &core_mesh.ExternalServiceResource{
			Meta: &test_model.ResourceMeta{Mesh: meshName, Name: "external-svc"},
			Spec: &mesh_proto.ExternalService{
				Networking: &mesh_proto.ExternalService_Networking{
					Address: "httpbin.org:80",
				},
				Tags: map[string]string{mesh_proto.ServiceTag: "external-svc"},
			},
		}
		Expect(resourceStore.Create(context.Background(), externalService, store.CreateByKey("external-svc", meshName))).To(Succeed())

		// and a builtin and a delegated gateway dataplane, neither of which is a regular service
		builtinGatewayBuilder := builders.Dataplane().
			WithMesh(meshName).
			WithName("gateway-builtin-dp").
			WithAddress("127.0.0.1").
			WithBuiltInGateway("gateway-builtin")
		Expect(builtinGatewayBuilder.Create(resourceStore)).To(Succeed())

		delegatedGatewayBuilder := builders.Dataplane().
			WithMesh(meshName).
			WithName("gateway-delegated-dp").
			WithAddress("127.0.0.1").
			WithBuiltInGateway("gateway-delegated")
		delegatedGateway := delegatedGatewayBuilder.Build()
		delegatedGateway.Spec.Networking.Gateway.Type = mesh_proto.Dataplane_Networking_Gateway_DELEGATED
		Expect(resourceStore.Create(context.Background(), delegatedGateway, store.CreateByKey("gateway-delegated-dp", meshName))).To(Succeed())

		// when
		mc, err := meshContextBuilder.Build(context.Background(), meshName)
		Expect(err).ToNot(HaveOccurred())

		// then the MeshService-backed service is TLS ready
		msKey := destinationname.MustResolve(false, ms, ms.Spec.Ports[0])
		Expect(mc.ServicesInformation[msKey]).ToNot(BeNil())
		Expect(mc.ServicesInformation[msKey].TLSReadiness).To(BeTrue())

		// and the external service is unconditionally TLS ready
		Expect(mc.ServicesInformation["external-svc"]).ToNot(BeNil())
		Expect(mc.ServicesInformation["external-svc"].IsExternalService).To(BeTrue())
		Expect(mc.ServicesInformation["external-svc"].TLSReadiness).To(BeTrue())

		// and gateway dataplanes (builtin and delegated) never had ServiceInsight-backed
		// TLS readiness and still don't get a ServicesInformation entry of their own
		Expect(mc.ServicesInformation).ToNot(HaveKey("gateway-builtin"))
		Expect(mc.ServicesInformation).ToNot(HaveKey("gateway-delegated"))
	})

	It("does not mark services TLS ready when the mesh CA backend is not PERMISSIVE", func() {
		meshBuilder := builders.Mesh().
			WithBuiltinMTLSBackend("ca-1").
			WithEnabledMTLSBackend("ca-1")
		Expect(meshBuilder.Create(resourceStore)).To(Succeed())
		meshName := meshBuilder.Build().GetMeta().GetName()

		msBuilder := builders.MeshService().
			WithMesh(meshName).
			WithName("backend").
			WithDataplaneTagsSelectorKV(mesh_proto.ServiceTag, "backend").
			AddIntPort(80, 8080, core_meta.ProtocolHTTP).
			WithTLSStatus(meshservice_api.TLSReady)
		Expect(msBuilder.Create(resourceStore)).To(Succeed())
		ms := msBuilder.Build()

		mc, err := meshContextBuilder.Build(context.Background(), meshName)
		Expect(err).ToNot(HaveOccurred())

		msKey := destinationname.MustResolve(false, ms, ms.Spec.Ports[0])
		info, found := mc.ServicesInformation[msKey]
		if found {
			Expect(info.TLSReadiness).To(BeFalse())
		}
	})
})
