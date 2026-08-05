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
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

var _ = Describe("GenerateClusters", func() {
	// A proxy is given its own identity before the destination reports that it
	// can terminate TLS, so an outbound cluster must stay on plaintext until the
	// destination's MeshService is TLS Ready, otherwise every request sent in
	// the window between the two pushes is dropped.
	type testCase struct {
		tlsStatus   meshservice_api.TLSStatus
		zoneOrigin  bool
		expectMTLS  bool
		expectedSNI string
		expectedSAN string
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
			AddSpiffeIDIdentity("spiffe://default/backend").
			WithTLSStatus(given.tlsStatus).
			Build()

		meshCtx := xds_context.MeshContext{
			Resource: builders.Mesh().Build(),
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{ms}),
			},
			ServicesInformation: map[string]*xds_context.ServiceInformation{},
		}

		backendRef := resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: kri.WithSectionName(kri.From(ms), "http"),
			Weight:   100,
		})
		services := envoy_common.NewServicesAccumulator(nil)
		services.AddBackendRef(backendRef, policies_xds.NewClusterBuilder().WithService("backend").Build())

		proxy := xds_builders.Proxy().
			WithWorkloadIdentity(xds_builders.WorkloadIdentity()).
			WithDataplane(builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
			Build()

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
			Expect(sans).To(HaveLen(1))
			Expect(sans[0].GetMatcher().GetExact()).To(Equal(given.expectedSAN))
		},
		Entry("local destination not TLS ready", testCase{
			tlsStatus:  meshservice_api.TLSNotReady,
			zoneOrigin: true,
		}),
		Entry("local destination TLS ready", testCase{
			tlsStatus:   meshservice_api.TLSReady,
			zoneOrigin:  true,
			expectMTLS:  true,
			expectedSNI: "sni.msvc.default.zone-1.backend.http",
			expectedSAN: "spiffe://default/backend",
		}),
		Entry("synced destination is always reachable over TLS", testCase{
			tlsStatus:   meshservice_api.TLSNotReady,
			expectMTLS:  true,
			expectedSNI: "sni.msvc.default.zone-1.backend.http",
			expectedSAN: "spiffe://default/backend",
		}),
	)
})
