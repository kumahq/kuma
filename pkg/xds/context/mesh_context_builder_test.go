package context_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
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

// remoteMeshZoneAddress is the public address of the zone proxy in zone "east",
// as it arrives over KDS. On Kubernetes it's reconciled from the zone ingress
// Service, on Universal it's authored by the user.
const remoteMeshZoneAddress = `
type: MeshZoneAddress
name: zone-proxy-east
mesh: default
labels:
  kuma.io/zone: east
  kuma.io/origin: global
spec:
  address: 192.168.0.1
  port: 10001
`

// remoteMeshZoneAddressWithHostname is the same zone proxy address, published as a
// load balancer hostname instead of an IP, which is what the ingress Service
// reconciler produces on EKS.
const remoteMeshZoneAddressWithHostname = `
type: MeshZoneAddress
name: zone-proxy-east
mesh: default
labels:
  kuma.io/zone: east
  kuma.io/origin: global
spec:
  address: lb.example.com
  port: 10001
`

func endpointTargets(meshCtx *xds_context.MeshContext) []string {
	var targets []string
	for _, endpoints := range meshCtx.EndpointMap {
		for _, endpoint := range endpoints {
			targets = append(targets, endpoint.Target)
		}
	}
	return targets
}

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
        kuma.io/display-name: backend
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
        kuma.io/display-name: backend
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
        kuma.io/display-name: backend
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
        kuma.io/display-name: backend
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

	It("recomputes the mesh context when a remote MeshService and its MeshZoneAddress newly appear", func() {
		// given a mesh whose proxies get a workload identity, matching the e2e repro
		Expect(samples.MeshDefaultBuilder().Create(resourceStore)).To(Succeed())
		Expect(builders.MeshIdentity().Create(resourceStore)).To(Succeed())

		before, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(before.EndpointMap).To(BeEmpty())

		// when a remote zone's MeshZoneAddress (the public address of its zone
		// proxy) and its auto-generated MeshService arrive together, as they
		// would over KDS when a new zone joins
		Expect(test_store.LoadResources(context.Background(), resourceStore, remoteMeshZoneAddress)).To(Succeed())
		Expect(samples.MeshServiceSyncedBackendBuilder().Create(resourceStore)).To(Succeed())

		after, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", before)
		Expect(err).ToNot(HaveOccurred())

		// then the mesh context must be rebuilt, with the new remote endpoint present
		Expect(after.Hash).ToNot(Equal(before.Hash), "a newly synced remote MeshService+MeshZoneAddress must invalidate the mesh context")
		Expect(after).ToNot(BeIdenticalTo(before))
		Expect(after.EndpointMap).ToNot(BeEmpty(), "the remote MeshService should get an endpoint via the MeshZoneAddress")
	})

	It("recomputes the mesh context when the MeshZoneAddress hostname resolves to a new IP", func() {
		// given a MeshZoneAddress pointing at a load balancer hostname, as reconciled
		// from a LoadBalancer Service on EKS, where the hostname takes precedence
		resolvedIP := "10.0.0.1"
		builderWithMutableDNS := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			func(host string) ([]net.IP, error) {
				if ip := net.ParseIP(host); ip != nil {
					return []net.IP{ip}, nil
				}
				return []net.IP{net.ParseIP(resolvedIP)}, nil
			},
			"zone-1",
		)
		Expect(samples.MeshDefaultBuilder().Create(resourceStore)).To(Succeed())
		Expect(builders.MeshIdentity().Create(resourceStore)).To(Succeed())
		Expect(test_store.LoadResources(context.Background(), resourceStore, remoteMeshZoneAddressWithHostname)).To(Succeed())
		Expect(samples.MeshServiceSyncedBackendBuilder().Create(resourceStore)).To(Succeed())

		before, err := builderWithMutableDNS.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(endpointTargets(before)).To(ConsistOf("10.0.0.1"))

		// when the load balancer rotates to a different IP, without any resource changing
		resolvedIP = "10.0.0.2"

		// then the mesh context must be rebuilt so proxies stop dialing the stale address
		after, err := builderWithMutableDNS.BuildIfChanged(context.Background(), "default", before)
		Expect(err).ToNot(HaveOccurred())
		Expect(after.Hash).ToNot(Equal(before.Hash), "re-resolved MeshZoneAddress must invalidate the mesh context")
		Expect(endpointTargets(after)).To(ConsistOf("10.0.0.2"))
	})

	It("recomputes the mesh context through the staged arrival matching the real KDS sequence", func() {
		// given a mesh whose proxies get a workload identity
		Expect(samples.MeshDefaultBuilder().Create(resourceStore)).To(Succeed())
		Expect(builders.MeshIdentity().Create(resourceStore)).To(Succeed())

		ctx0, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())

		// stage 1: the remote zone's MeshZoneAddress arrives first, fully resolved
		// (matches the real KDS trace: the zone proxy address is observed complete from creationTime)
		Expect(test_store.LoadResources(context.Background(), resourceStore, remoteMeshZoneAddress)).To(Succeed())

		ctx1, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", ctx0)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx1.Hash).ToNot(Equal(ctx0.Hash), "MeshZoneAddress arrival alone must invalidate the mesh context")
		Expect(ctx1.EndpointMap).To(BeEmpty(), "no MeshService exists yet, so there's nothing to route to")

		// stage 2: the auto-generated MeshService is created next, WITHOUT a VIP yet
		// (matches the real KDS trace: MeshService created ~2s before "vips.allocator: allocating IP")
		Expect(samples.MeshServiceSyncedBackendBuilder().WithoutVIP().Create(resourceStore)).To(Succeed())

		ctx2, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", ctx1)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx2.Hash).ToNot(Equal(ctx1.Hash), "MeshService arrival must invalidate the mesh context")
		Expect(ctx2.EndpointMap).ToNot(BeEmpty(), "cross-zone routing goes through the MeshZoneAddress and does not require a VIP")

		// stage 3: the VIP is allocated a couple seconds later
		meshService := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), meshService, store.GetByKey(samples.MeshServiceSyncedBackendBuilder().Build().GetMeta().GetName(), "default"))).To(Succeed())
		meshService.Status.VIPs = []meshservice_api.VIP{{IP: "240.0.0.3"}}
		Expect(resourceStore.Update(context.Background(), meshService, store.UpdateWithLabels(meshService.GetMeta().GetLabels()))).To(Succeed())

		ctx3, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", ctx2)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx3.Hash).ToNot(Equal(ctx2.Hash), "VIP allocation must invalidate the mesh context")
	})

	It("resolves a Dataplane address that is a DNS name", func() {
		// given a builder whose lookup resolves a hostname, as it would on AWS ECS
		// where the Dataplane is created before the container gets its IP
		builderWithDNS := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			func(host string) ([]net.IP, error) {
				switch host {
				case "backend.dns.name":
					return []net.IP{net.ParseIP("192.168.0.10")}, nil
				default:
					return []net.IP{net.ParseIP(host)}, nil
				}
			},
			"zone-1",
		)

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: backend.dns.name
  inbound:
    - port: 8080
---
type: MeshService
name: backend
mesh: mesh-1
spec:
  selector:
    dataplaneRef:
      name: dp-1
  ports:
  - port: 8080
    targetPort: 8080
    appProtocol: tcp
`)).To(Succeed())

		// when
		meshCtx, err := builderWithDNS.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		// then the dataplane stored in the mesh context is resolved
		dataplanes := meshCtx.Resources.Dataplanes().Items
		Expect(dataplanes).To(HaveLen(1))
		Expect(dataplanes[0].Spec.GetNetworking().GetAddress()).To(Equal("192.168.0.10"))

		// and so is the endpoint Envoy EDS gets, since it only accepts IPs
		meshService := meshCtx.Resources.MeshServices().Items[0]
		endpoints := meshCtx.EndpointMap[destinationname.MustResolve(meshService, meshService.Spec.Ports[0])]
		Expect(endpoints).To(HaveLen(1))
		Expect(endpoints[0].Target).To(Equal("192.168.0.10"))
	})

	It("returns an error instead of panicking when listing resources fails", func() {
		// given a mesh exists but listing any other resource fails (e.g. DB connection lost)
		Expect(samples.MeshDefaultBuilder().Create(resourceStore)).To(Succeed())
		builder := xds_context.NewMeshContextBuilder(
			&failingListManager{ReadOnlyResourceManager: resourceStore},
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
		)

		// when building the base mesh context
		_, err := builder.BuildBaseMeshContextIfChanged(context.Background(), "default", nil)

		// then the error is surfaced rather than causing a nil-pointer panic
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to build base mesh context"))
	})
})

// failingListManager delegates Get to a real manager but fails every List, simulating a store error.
type failingListManager struct {
	manager.ReadOnlyResourceManager
}

func (f *failingListManager) List(context.Context, core_model.ResourceList, ...store.ListOptionsFunc) error {
	return errors.New("store unavailable")
}

var _ = Describe("EndpointMap", func() {
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
		)
	})

	It("resolves protocol and marks external services, skipping gateway dataplanes", func() {
		// given
		meshBuilder := builders.Mesh()
		Expect(meshBuilder.Create(resourceStore)).To(Succeed())
		meshName := meshBuilder.Build().GetMeta().GetName()

		// and a MeshService-backed service with no Dataplane behind it, so it
		// contributes no endpoints
		msBuilder := builders.MeshService().
			WithMesh(meshName).
			WithName("backend").
			WithDataplaneLabelsSelectorKV("kuma.io/display-name", "backend").
			AddIntPort(80, 8080, core_meta.ProtocolHTTP)
		Expect(msBuilder.Create(resourceStore)).To(Succeed())
		ms := msBuilder.Build()

		// and a MeshExternalService
		externalService := &meshexternalservice_api.MeshExternalServiceResource{
			Meta: &test_model.ResourceMeta{Mesh: meshName, Name: "external-svc"},
			Spec: &meshexternalservice_api.MeshExternalService{
				Match: meshexternalservice_api.Match{
					Type:     meshexternalservice_api.HostnameGeneratorType,
					Port:     80,
					Protocol: core_meta.ProtocolHTTP,
				},
				Endpoints: &[]meshexternalservice_api.Endpoint{
					{
						Address: "httpbin.org",
						Port:    80,
					},
				},
			},
			Status: &meshexternalservice_api.MeshExternalServiceStatus{},
		}
		Expect(resourceStore.Create(context.Background(), externalService, store.CreateByKey("external-svc", meshName))).To(Succeed())

		// and a ready zone egress listener so MeshExternalService endpoints are materialized
		zoneEgress := &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Mesh: meshName, Name: "zone-egress-dp"},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.10",
					Listeners: []*mesh_proto.Dataplane_Networking_Listener{
						{
							Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
							Address: "127.0.0.10",
							Port:    10002,
							State:   mesh_proto.Dataplane_Networking_Listener_Ready,
						},
					},
				},
			},
		}
		Expect(resourceStore.Create(context.Background(), zoneEgress, store.CreateByKey("zone-egress-dp", meshName))).To(Succeed())

		// and a delegated gateway dataplane, which is not a regular service
		delegatedGatewayBuilder := builders.Dataplane().
			WithMesh(meshName).
			WithName("gateway-delegated-dp").
			WithAddress("127.0.0.1").
			WithDelegatedGateway()
		Expect(delegatedGatewayBuilder.Create(resourceStore)).To(Succeed())

		// when
		mc, err := meshContextBuilder.Build(context.Background(), meshName)
		Expect(err).ToNot(HaveOccurred())

		// then a MeshService with no Dataplane behind it contributes no endpoints
		msKey := destinationname.MustResolve(ms, ms.Spec.Ports[0])
		Expect(mc.EndpointMap).ToNot(HaveKey(msKey))

		// and the external service resolves to the egress, carrying the protocol
		// declared on the resource
		esKey := destinationname.MustResolve(externalService, externalService.Spec.Match)
		Expect(mc.EndpointMap[esKey]).ToNot(BeEmpty())
		Expect(mc.EndpointMap[esKey][0].IsExternalService()).To(BeTrue())
		Expect(mc.EndpointMap[esKey][0].ExternalService.Protocol).To(Equal(core_meta.ProtocolHTTP))

		// and gateway dataplanes are not destinations
		Expect(mc.EndpointMap).ToNot(HaveKey("gateway-delegated"))
	})
})
