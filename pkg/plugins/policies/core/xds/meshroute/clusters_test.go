package meshroute_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v2/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/v2/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
)

var _ = Describe("SniForBackendRef", func() {
	DescribeTable("returns SNI built from resolved port",
		func(sectionName string) {
			ms := builders.MeshService().
				WithName("backend").
				WithMesh("default").
				AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
				Build()

			port, ok := ms.FindPortByName(sectionName)
			Expect(ok).To(BeTrue())

			id := kri.WithSectionName(kri.From(ms), sectionName)
			ref := &resolve.RealResourceBackendRef{Resource: id}

			sni := meshroute.SniForBackendRef(ref, ms, port, "")

			Expect(sni).NotTo(BeEmpty())
			Expect(sni).To(ContainSubstring(".8080."))
		},
		Entry("by port name", "http"),
		Entry("by port value", "8080"),
	)

	It("uses SNIName for MeshService destination", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
			Build()

		port, ok := ms.FindPortByName("http")
		Expect(ok).To(BeTrue())

		id := kri.WithSectionName(kri.From(ms), "http")
		id.ResourceType = meshservice_api.MeshServiceType
		ref := &resolve.RealResourceBackendRef{Resource: id}

		sni := meshroute.SniForBackendRef(ref, ms, port, "kuma-system")

		Expect(sni).To(ContainSubstring(ms.SNIName("kuma-system")))
	})
})

// The egress pool a MeshExternalService is reached through decides how that connection is
// secured, and the cluster has to be emitted whichever pool is in use.
var _ = Describe("GenerateClusters for MeshExternalService", func() {
	const egressSAN = "spiffe://default.zone-1.mesh.local/ns/kuma-system/sa/kuma-default-egress"

	mes := builders.MeshExternalService().WithName("httpbin").Build()
	serviceName := kri.From(mes).String()

	workloadIdentity := &core_xds.WorkloadIdentity{
		IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return bldrs_tls.SdsSecretConfigSource(
				"identity_cert:secret:default",
				bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
			)
		},
	}

	buildInput := func(
		mesh *core_mesh.MeshResource,
		workloadIdentity *core_xds.WorkloadIdentity,
		zoneEgresses []core_xds.ZoneEgressInstance,
	) (*core_xds.Proxy, xds_context.MeshContext, envoy_common.Services) {
		proxy := &core_xds.Proxy{
			APIVersion:       envoy_common.APIV3,
			Dataplane:        samples.DataplaneBackend(),
			Metadata:         &core_xds.DataplaneMetadata{},
			WorkloadIdentity: workloadIdentity,
			SecretsTracker:   envoy_common.NewSecretsTracker(core_model.DefaultMesh, nil),
		}

		var endpoints []core_xds.Endpoint
		for _, ze := range zoneEgresses {
			endpoints = append(endpoints, core_xds.Endpoint{
				Target:          ze.Address,
				Port:            ze.Port,
				ExternalService: &core_xds.ExternalService{OwnerResource: kri.From(mes)},
			})
		}

		meshCtx := xds_context.MeshContext{
			Resource: mesh,
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{mes}),
			},
			ServicesInformation: map[string]*xds_context.ServiceInformation{
				serviceName: {IsExternalService: true},
			},
			EndpointMap:  core_xds.EndpointMap{serviceName: endpoints},
			ZoneEgresses: zoneEgresses,
		}

		ref := resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: kri.WithSectionName(kri.From(mes), "9000"),
			Weight:   1,
		})

		sa := envoy_common.NewServicesAccumulator(nil)
		sa.AddBackendRef(ref, envoy_common.NewCluster(
			envoy_common.WithName(serviceName),
			envoy_common.WithService(serviceName),
			envoy_common.WithExternalService(true),
		))

		return proxy, meshCtx, sa.Services()
	}

	// Endpoints are built once per egress instance, so a destination reached through a pool
	// always has them: a bare cluster would send cleartext into a listener terminating TLS.
	legacyPool := []core_xds.ZoneEgressInstance{{Address: "192.168.0.1", Port: 10002}}
	meshScopedPool := []core_xds.ZoneEgressInstance{{Address: "192.168.0.2", Port: 10002, SAN: egressSAN}}

	transportSocketOf := func(rs *core_xds.ResourceSet) *envoy_cluster.Cluster {
		clusters := rs.Resources(envoy_resource.ClusterType)
		ExpectWithOffset(1, clusters).To(HaveKey(serviceName))
		return clusters[serviceName].Resource.(*envoy_cluster.Cluster)
	}

	It("verifies the mesh CA identity when the pool is legacy", func() {
		// given a proxy that has a workload identity and a pool that has none
		proxy, meshCtx, services := buildInput(samples.MeshMTLS(), workloadIdentity, legacyPool)

		// when
		rs, err := meshroute.GenerateClusters(proxy, meshCtx, services, "kuma-system")

		// then
		Expect(err).ToNot(HaveOccurred())
		cluster := transportSocketOf(rs)
		Expect(cluster.TransportSocket).ToNot(BeNil())
		tlsCtx := &envoy_tls.UpstreamTlsContext{}
		Expect(util_proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), tlsCtx)).To(Succeed())
		Expect(tlsCtx.GetSni()).To(ContainSubstring("httpbin"))
	})

	It("verifies the egress SPIFFE ID when the pool is mesh-scoped", func() {
		// given
		proxy, meshCtx, services := buildInput(samples.MeshMTLS(), workloadIdentity, meshScopedPool)

		// when
		rs, err := meshroute.GenerateClusters(proxy, meshCtx, services, "kuma-system")

		// then
		Expect(err).ToNot(HaveOccurred())
		cluster := transportSocketOf(rs)
		Expect(cluster.TransportSocket).ToNot(BeNil())
		tlsCtx := &envoy_tls.UpstreamTlsContext{}
		Expect(util_proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), tlsCtx)).To(Succeed())
		matchers := tlsCtx.GetCommonTlsContext().GetCombinedValidationContext().GetDefaultValidationContext().GetMatchTypedSubjectAltNames()
		Expect(matchers).To(HaveLen(1))
		Expect(matchers[0].GetMatcher().GetExact()).To(Equal(egressSAN))
	})

	// Identity is resolved per proxy and the pool per mesh, so they disagree while issuance
	// catches up. Nothing this proxy is served can work, but the cluster still has to exist.
	It("still emits the cluster when the proxy has no identity for a mesh-scoped pool", func() {
		// given
		proxy, meshCtx, services := buildInput(samples.MeshMTLS(), nil, meshScopedPool)

		// when
		rs, err := meshroute.GenerateClusters(proxy, meshCtx, services, "kuma-system")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(transportSocketOf(rs)).ToNot(BeNil())
	})

	// Zone egress needs mTLS to work at all, and there are no mesh secrets to reference
	// without it, so this is the only case that comes out with no transport socket.
	It("emits a bare cluster only when the mesh has no mTLS", func() {
		// given
		proxy, meshCtx, services := buildInput(samples.MeshDefault(), workloadIdentity, legacyPool)

		// when
		rs, err := meshroute.GenerateClusters(proxy, meshCtx, services, "kuma-system")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(transportSocketOf(rs).TransportSocket).To(BeNil())
	})

	DescribeTable("emits a cluster for every load assignment",
		func(zoneEgresses []core_xds.ZoneEgressInstance, workloadIdentity *core_xds.WorkloadIdentity) {
			// given
			proxy, meshCtx, services := buildInput(samples.MeshMTLS(), workloadIdentity, zoneEgresses)
			xdsCtx := xds_context.Context{
				Mesh: meshCtx,
				ControlPlane: &xds_context.ControlPlaneContext{
					CLACache: &test_xds.DummyCLACache{OutboundTargets: meshCtx.EndpointMap},
				},
			}

			// when
			clusters, err := meshroute.GenerateClusters(proxy, meshCtx, services, "kuma-system")
			Expect(err).ToNot(HaveOccurred())
			endpoints, err := meshroute.GenerateEndpoints(proxy, xdsCtx, services)
			Expect(err).ToNot(HaveOccurred())

			// then every emitted load assignment has a cluster to belong to
			Expect(clusters.Resources(envoy_resource.ClusterType)).To(
				HaveLen(len(endpoints.Resources(envoy_resource.EndpointType))))
		},
		Entry("legacy pool", legacyPool, workloadIdentity),
		Entry("mesh-scoped pool", meshScopedPool, workloadIdentity),
		Entry("mesh-scoped pool, proxy without identity", meshScopedPool, nil),
	)
})
