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
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
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

// During the MeshIdentity migration a MeshService advertises both a ServiceTag identity
// (legacy mTLS) and a SpiffeID identity (new MeshIdentity). A proxy that already has a
// workload identity (includeSpiffeID) must accept the legacy certificate as well, so the
// ServiceTag has to be rendered as spiffe://<mesh>/<service>. MeshService and
// MeshMultiZoneService backendRefs must produce the same SANs for the same backends.
var _ = Describe("Identities", func() {
	const (
		legacySAN     = "spiffe://default/echo-http"
		universalSAN  = "spiffe://default.universal-zone.mesh.local/workload/echo-http"
		kubernetesSAN = "spiffe://default.k8s-zone.mesh.local/ns/echo/sa/echo-http"
		serviceTag    = "echo-http"
	)

	msUniversal := func() *meshservice_api.MeshServiceResource {
		ms := builders.MeshService().
			WithName("echo-http").
			WithMesh("default").
			WithZone("universal-zone").
			AddIntPortWithName(80, 80, core_meta.ProtocolHTTP, "http").
			AddServiceTagIdentity(serviceTag).
			Build()
		ms.Spec.Identities = pointer.To(append(pointer.Deref(ms.Spec.Identities), meshservice_api.MeshServiceIdentity{
			Type:  meshservice_api.MeshServiceIdentitySpiffeIDType,
			Value: universalSAN,
		}))
		return ms
	}

	msKubernetes := func() *meshservice_api.MeshServiceResource {
		ms := builders.MeshService().
			WithName("echo-http").
			WithMesh("default").
			WithZone("k8s-zone").
			AddIntPortWithName(80, 80, core_meta.ProtocolHTTP, "http").
			AddServiceTagIdentity(serviceTag).
			Build()
		ms.Spec.Identities = pointer.To(append(pointer.Deref(ms.Spec.Identities), meshservice_api.MeshServiceIdentity{
			Type:  meshservice_api.MeshServiceIdentitySpiffeIDType,
			Value: kubernetesSAN,
		}))
		return ms
	}

	meshContextWith := func(resources ...core_model.Resource) xds_context.MeshContext {
		return xds_context.MeshContext{
			Resource: samples.MeshDefault(),
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex(resources),
			},
		}
	}

	DescribeTable("produces the same SANs for MeshService and MeshMultiZoneService",
		func(includeSpiffeID bool, expected []string) {
			// given a MeshMultiZoneService matching the universal MeshService
			ms := msUniversal()
			mzms := builders.MeshMultiZoneService().
				WithName("echo-http-uz").
				WithMesh("default").
				WithServiceLabelSelector(map[string]string{"kuma.io/service-name": "echo-http"}).
				AddMatchedMeshServiceName(kri.From(ms)).
				AddIntPort(80, core_meta.ProtocolHTTP).
				Build()
			meshCtx := meshContextWith(ms, mzms)

			// when
			msSANs := meshroute.Identities(&resolve.RealResourceBackendRef{Resource: kri.From(ms)}, meshCtx, includeSpiffeID)
			mzmsSANs := meshroute.Identities(&resolve.RealResourceBackendRef{Resource: kri.From(mzms)}, meshCtx, includeSpiffeID)

			// then both paths accept the same identities
			Expect(msSANs).To(Equal(expected))
			Expect(mzmsSANs).To(Equal(expected))
		},
		Entry("proxy with a workload identity accepts the legacy SAN (migration stage)",
			true, []string{universalSAN, legacySAN}),
		Entry("proxy without a workload identity keeps the raw service tag",
			false, []string{serviceTag, universalSAN}),
	)

	It("deduplicates the legacy SAN shared by matched services from different zones", func() {
		// given matched MeshServices in two zones presenting the same legacy identity
		msUz := msUniversal()
		msK8s := msKubernetes()
		mzms := builders.MeshMultiZoneService().
			WithName("echo-http-all").
			WithMesh("default").
			WithServiceLabelSelector(map[string]string{"kuma.io/service-name": "echo-http"}).
			AddMatchedMeshServiceName(kri.From(msUz)).
			AddMatchedMeshServiceName(kri.From(msK8s)).
			AddIntPort(80, core_meta.ProtocolHTTP).
			Build()
		meshCtx := meshContextWith(msUz, msK8s, mzms)

		// when
		sans := meshroute.Identities(&resolve.RealResourceBackendRef{Resource: kri.From(mzms)}, meshCtx, true)

		// then
		Expect(sans).To(Equal([]string{kubernetesSAN, universalSAN, legacySAN}))
	})

	It("skips matched MeshServices missing from the mesh context", func() {
		// given a MeshMultiZoneService whose matched MeshService is not synced yet
		ms := msUniversal()
		mzms := builders.MeshMultiZoneService().
			WithName("echo-http-uz").
			WithMesh("default").
			WithServiceLabelSelector(map[string]string{"kuma.io/service-name": "echo-http"}).
			AddMatchedMeshServiceName(kri.From(ms)).
			AddIntPort(80, core_meta.ProtocolHTTP).
			Build()
		meshCtx := meshContextWith(mzms)

		// when
		sans := meshroute.Identities(&resolve.RealResourceBackendRef{Resource: kri.From(mzms)}, meshCtx, true)

		// then
		Expect(sans).To(BeEmpty())
	})
})

// Identities may be unknown yet, e.g. while a matched MeshService is still syncing.
// An empty SAN list must not turn into "accept any certificate chaining to the trust
// bundle" — fail closed on the mesh SPIFFE prefix like the legacy mTLS path does.
var _ = Describe("UpstreamTLSContext", func() {
	workloadIdentity := &core_xds.WorkloadIdentity{
		IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return bldrs_tls.SdsSecretConfigSource(
				"identity_cert:secret:default",
				bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
			)
		},
	}

	proxy := func() *core_xds.Proxy {
		return &core_xds.Proxy{
			APIVersion:       envoy_common.APIV3,
			Dataplane:        samples.DataplaneBackend(),
			Metadata:         &core_xds.DataplaneMetadata{},
			WorkloadIdentity: workloadIdentity,
			SecretsTracker:   envoy_common.NewSecretsTracker(core_model.DefaultMesh, nil),
		}
	}

	sanMatchers := func(ctx *envoy_tls.UpstreamTlsContext) []*envoy_tls.SubjectAltNameMatcher {
		return ctx.GetCommonTlsContext().
			GetCombinedValidationContext().
			GetDefaultValidationContext().
			GetMatchTypedSubjectAltNames()
	}

	It("falls back to the mesh SPIFFE prefix when no identities are known", func() {
		// when
		ctx, err := meshroute.UpstreamTLSContext(proxy(), "backend", nil)

		// then
		Expect(err).ToNot(HaveOccurred())
		matchers := sanMatchers(ctx)
		Expect(matchers).To(HaveLen(1))
		Expect(matchers[0].GetSanType()).To(Equal(envoy_tls.SubjectAltNameMatcher_URI))
		Expect(matchers[0].GetMatcher().GetPrefix()).To(Equal("spiffe://default/"))
	})

	It("uses exact matchers for the given identities", func() {
		// when
		ctx, err := meshroute.UpstreamTLSContext(proxy(), "backend", []string{
			"spiffe://default/backend",
			"spiffe://default.universal-zone.mesh.local/workload/backend",
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		matchers := sanMatchers(ctx)
		Expect(matchers).To(HaveLen(2))
		Expect(matchers[0].GetMatcher().GetExact()).To(Equal("spiffe://default/backend"))
		Expect(matchers[1].GetMatcher().GetExact()).To(Equal("spiffe://default.universal-zone.mesh.local/workload/backend"))
	})
})
