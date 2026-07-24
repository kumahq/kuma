package hooks_test

import (
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s/xds/hooks"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var _ = Describe("ApiServerBypass", func() {
	It("should generate configuration for API Server bypass", func() {
		// given
		hook := hooks.NewApiServerBypass("1.1.1.1", 9090)
		rs := xds.NewResourceSet()
		ctx := xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &mesh.MeshResource{
					Spec: &mesh_proto.Mesh{},
				},
			},
		}
		proxy := &xds.Proxy{
			APIVersion: envoy.APIV3,
			Dataplane:  &mesh.DataplaneResource{},
		}

		// when
		Expect(hook.Modify(rs, ctx, proxy)).To(Succeed())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "api-server-bypass.yaml")))
	})

	It("should generate configuration for API Server bypass with unified naming", func() {
		// given
		hook := hooks.NewApiServerBypass("1.1.1.1", 9090)
		rs := xds.NewResourceSet()
		ctx := xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &mesh.MeshResource{
					Spec: &mesh_proto.Mesh{},
				},
			},
		}
		proxy := &xds.Proxy{
			APIVersion: envoy.APIV3,
			Dataplane:  &mesh.DataplaneResource{},
			Metadata: &xds.DataplaneMetadata{
				Features: map[string]bool{
					types.FeatureUnifiedResourceNaming: true,
				},
			},
		}

		// when
		Expect(hook.Modify(rs, ctx, proxy)).To(Succeed())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "api-server-bypass-unified-naming.yaml")))
	})

	It("should not generate configuration when the default outbound passthrough cluster is already present", func() {
		// given
		hook := hooks.NewApiServerBypass("1.1.1.1", 9090)
		rs := xds.NewResourceSet()
		passthroughCluster, err := envoy_clusters.NewClusterBuilder(envoy.APIV3, metadata.TransparentOutboundNameIPv4).
			Configure(envoy_clusters.PassThroughCluster()).
			Build()
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&xds.Resource{
			Name:     passthroughCluster.GetName(),
			Origin:   metadata.OriginTransparent,
			Resource: passthroughCluster,
		})
		ctx := xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &mesh.MeshResource{
					Spec: &mesh_proto.Mesh{},
				},
			},
		}
		proxy := &xds.Proxy{
			APIVersion: envoy.APIV3,
			Dataplane:  &mesh.DataplaneResource{},
		}

		// when
		Expect(hook.Modify(rs, ctx, proxy)).To(Succeed())

		// then no additional bypass listener/cluster was generated
		Expect(rs.Resources(envoy_resource.ClusterType)).To(HaveLen(1))
		Expect(rs.ResourceTypes()).ToNot(ContainElement(envoy_resource.ListenerType))
	})
})
