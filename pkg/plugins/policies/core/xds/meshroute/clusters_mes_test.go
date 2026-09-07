package meshroute_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
)

// GenerateEndpoints always emits a ClusterLoadAssignment for a
// MeshExternalService destination. If GenerateClusters skips the matching
// cluster the snapshot becomes inconsistent and the proxy is left without any
// configuration at all, so the cluster has to be emitted even when there is no
// zone egress to route through.
var _ = Describe("GenerateClusters for MeshExternalService", func() {
	buildInput := func(zoneEgresses []core_xds.ZoneEgressInstance) (*core_xds.Proxy, xds_context.MeshContext, envoy_common.Services) {
		mes := builders.MeshExternalService().WithName("httpbin").Build()
		mesKRI := kri.From(mes)

		proxy := &core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
			Dataplane:  samples.DataplaneBackend(),
			Metadata:   &core_xds.DataplaneMetadata{},
			WorkloadIdentity: &core_xds.WorkloadIdentity{
				KRI: kri.Identifier{},
			},
			SecretsTracker: envoy_common.NewSecretsTracker(core_model.DefaultMesh, nil),
		}

		serviceName := mesKRI.String()
		meshCtx := xds_context.MeshContext{
			Resource: samples.MeshDefault(),
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{mes}),
			},
			ServicesInformation: map[string]*xds_context.ServiceInformation{
				serviceName: {IsExternalService: true},
			},
			EndpointMap: core_xds.EndpointMap{
				serviceName: []core_xds.Endpoint{{
					Target: "192.168.0.1",
					Port:   27017,
					ExternalService: &core_xds.ExternalService{
						OwnerResource: mesKRI,
					},
				}},
			},
			ZoneEgresses: zoneEgresses,
		}

		ref := resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: kri.WithSectionName(mesKRI, "9000"),
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

	It("emits the cluster even when there is no zone egress", func() {
		proxy, meshCtx, services := buildInput(nil)

		rs, err := meshroute.GenerateClusters(proxy, meshCtx, services, "kuma-system")

		Expect(err).ToNot(HaveOccurred())
		// Without this the load assignment emitted by GenerateEndpoints would
		// have no cluster, and snapshot.Consistent() would reject the whole
		// configuration.
		clusters := rs.Resources(envoy_resource.ClusterType)
		Expect(clusters).To(HaveLen(1))
		Expect(clusters).To(HaveKey(kri.From(
			builders.MeshExternalService().WithName("httpbin").Build()).String()))
	})
})
