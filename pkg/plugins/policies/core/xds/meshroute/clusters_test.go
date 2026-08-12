package meshroute_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

var _ = Describe("SNIForRealResource", func() {
	DescribeTable("returns KRI SNI built from the resolved port name",
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

			sni, ok := meshroute.SNIForRealResource(ref, port)

			Expect(ok).To(BeTrue())
			Expect(sni).To(Equal("sni.msvc.default.backend.http"))
		},
		Entry("by port name", "http"),
		Entry("by port value", "8080"),
	)

	It("skips invalid destination KRI SNI", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "HTTP").
			Build()

		port, ok := ms.FindPortByName("HTTP")
		Expect(ok).To(BeTrue())

		ref := &resolve.RealResourceBackendRef{Resource: kri.WithSectionName(kri.From(ms), "HTTP")}

		_, valid := meshroute.SNIForRealResource(ref, port)
		Expect(valid).To(BeFalse())
	})
})

var _ = Describe("GenerateClusters", func() {
	// A proxy is given its own identity before the destination reports that it
	// can terminate TLS, so an outbound cluster must stay on plaintext until the
	// destination's MeshService is TLS Ready, otherwise every request sent in
	// the window between the two pushes is dropped.
	type testCase struct {
		tlsStatus    meshservice_api.TLSStatus
		zoneOrigin   bool
		noIdentity   bool
		expectMTLS   bool
		expectedSNI  string
		expectedSANs []string
	}

	buildCluster := func(given testCase) *envoy_cluster.Cluster {
		labels := map[string]string{mesh_proto.ZoneTag: "zone-1"}
		if !given.zoneOrigin {
			labels[mesh_proto.ResourceOriginLabel] = string(mesh_proto.GlobalResourceOrigin)
		}
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			WithLabels(labels).
			AddIntPortWithName(80, 8080, core_meta.ProtocolHTTP, "http").
			AddSpiffeIDIdentity("spiffe://default.zone-1.mesh.local/workload/backend").
			WithTLSStatus(given.tlsStatus).
			Build()

		meshCtx := xds_context.MeshContext{
			Resource: builders.Mesh().Build(),
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{ms}),
			},
		}

		backendRef := resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: kri.WithSectionName(kri.From(ms), "http"),
			Weight:   100,
		})
		services := envoy_common.NewServicesAccumulator()
		services.AddBackendRef(backendRef, policies_xds.NewClusterBuilder().WithService("backend").Build())

		proxyBuilder := xds_builders.Proxy().
			WithDataplane(builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"))
		if !given.noIdentity {
			proxyBuilder = proxyBuilder.WithWorkloadIdentity(xds_builders.WorkloadIdentity())
		}
		proxy := proxyBuilder.Build()

		rs, err := meshroute.GenerateClusters(proxy, meshCtx, services.Services())
		Expect(err).ToNot(HaveOccurred())

		clusters := rs.Resources(envoy_resource.ClusterType)
		Expect(clusters).To(HaveLen(1))
		cluster, ok := clusters["backend"].Resource.(*envoy_cluster.Cluster)
		Expect(ok).To(BeTrue())
		return cluster
	}

	DescribeTable("should configure upstream TLS only once the destination is TLS ready",
		func(given testCase) {
			cluster := buildCluster(given)

			if !given.expectMTLS {
				Expect(cluster.TransportSocket).To(BeNil())
				return
			}

			Expect(cluster.TransportSocket).ToNot(BeNil())
			upstreamCtx := &envoy_tls.UpstreamTlsContext{}
			Expect(util_proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), upstreamCtx)).To(Succeed())
			Expect(upstreamCtx.Sni).To(Equal(given.expectedSNI))

			sans := upstreamCtx.GetCommonTlsContext().GetCombinedValidationContext().GetDefaultValidationContext().GetMatchTypedSubjectAltNames()
			var exacts []string
			for _, san := range sans {
				exacts = append(exacts, san.GetMatcher().GetExact())
			}
			Expect(exacts).To(Equal(given.expectedSANs))
		},
		Entry("local destination not TLS ready", testCase{
			tlsStatus:  meshservice_api.TLSNotReady,
			zoneOrigin: true,
		}),
		Entry("local destination TLS ready", testCase{
			tlsStatus:    meshservice_api.TLSReady,
			zoneOrigin:   true,
			expectMTLS:   true,
			expectedSNI:  "sni.msvc.default.zone-1.backend.http",
			expectedSANs: []string{"spiffe://default.zone-1.mesh.local/workload/backend"},
		}),
		Entry("synced destination is always reachable over TLS", testCase{
			tlsStatus:    meshservice_api.TLSNotReady,
			expectMTLS:   true,
			expectedSNI:  "sni.msvc.default.zone-1.backend.http",
			expectedSANs: []string{"spiffe://default.zone-1.mesh.local/workload/backend"},
		}),
		Entry("proxy without a workload identity cannot originate mTLS", testCase{
			tlsStatus:  meshservice_api.TLSReady,
			zoneOrigin: true,
			noIdentity: true,
		}),
	)

	It("uses KRI SNI for MeshExternalService with WorkloadIdentity", func() {
		mes := builders.MeshExternalService().
			WithName("external-backend").
			WithMesh("default").
			WithKumaVIP("242.0.0.1").
			Build()

		meshCtx := xds_context.MeshContext{
			Resource: builders.Mesh().Build(),
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{mes}),
			},
			ZoneEgresses: []core_xds.ZoneEgressInstance{{
				Address: "10.0.0.1",
				Port:    10002,
				SAN:     "spiffe://default/zone-egress",
			}},
		}

		mesKRI := kri.WithSectionName(kri.From(mes), "9000")
		backendRef := resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: mesKRI,
			Weight:   100,
		})
		services := envoy_common.NewServicesAccumulator()
		services.AddBackendRef(backendRef, policies_xds.NewClusterBuilder().
			WithService(mesKRI.String()).
			WithExternalService(true).
			Build())

		rs, err := meshroute.GenerateClusters(
			xds_builders.Proxy().
				WithDataplane(builders.Dataplane().
					WithName("web-01").
					WithAddress("192.168.0.2").
					WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
				WithWorkloadIdentity(&core_xds.WorkloadIdentity{
					IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
						return bldrs_tls.SdsSecretConfigSource(
							"identity_cert:secret:default",
							bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
						)
					},
				}).
				Build(),
			meshCtx,
			services.Services(),
		)
		Expect(err).ToNot(HaveOccurred())

		clusters := rs.Resources(envoy_resource.ClusterType)
		Expect(clusters).To(HaveLen(1))
		cluster, ok := clusters[mesKRI.String()].Resource.(*envoy_cluster.Cluster)
		Expect(ok).To(BeTrue())
		Expect(cluster.TransportSocket).ToNot(BeNil())

		upstreamCtx := &envoy_tls.UpstreamTlsContext{}
		Expect(util_proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), upstreamCtx)).To(Succeed())
		Expect(upstreamCtx.Sni).To(Equal("sni.extsvc.default.external-backend.9000"))
	})

	It("uses KRI SNI for MeshMultiZoneService with WorkloadIdentity", func() {
		mzms := builders.MeshMultiZoneService().
			WithName("backend").
			WithMesh("default").
			WithServiceLabelSelector(map[string]string{"app": "backend"}).
			AddIntPortWithName(8080, core_meta.ProtocolHTTP, "http").
			Build()

		meshCtx := xds_context.MeshContext{
			Resource: builders.Mesh().Build(),
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{mzms}),
			},
		}

		backendRef := resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: kri.WithSectionName(kri.From(mzms), "8080"),
			Weight:   100,
		})
		services := envoy_common.NewServicesAccumulator()
		services.AddBackendRef(backendRef, policies_xds.NewClusterBuilder().WithService("backend").Build())

		rs, err := meshroute.GenerateClusters(
			xds_builders.Proxy().
				WithDataplane(builders.Dataplane().
					WithName("web-01").
					WithAddress("192.168.0.2").
					WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
				WithWorkloadIdentity(&core_xds.WorkloadIdentity{
					IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
						return bldrs_tls.SdsSecretConfigSource(
							"identity_cert:secret:default",
							bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
						)
					},
				}).
				Build(),
			meshCtx,
			services.Services(),
		)
		Expect(err).ToNot(HaveOccurred())

		clusters := rs.Resources(envoy_resource.ClusterType)
		Expect(clusters).To(HaveLen(1))
		cluster, ok := clusters["backend"].Resource.(*envoy_cluster.Cluster)
		Expect(ok).To(BeTrue())
		Expect(cluster.TransportSocket).ToNot(BeNil())

		upstreamCtx := &envoy_tls.UpstreamTlsContext{}
		Expect(util_proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), upstreamCtx)).To(Succeed())
		Expect(upstreamCtx.Sni).To(Equal("sni.mzsvc.default.backend.http"))
		Expect(cluster.TransportSocketMatches).To(BeEmpty())
	})
})
