package matchers_test

import (
	"fmt"
	"testing"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func BenchmarkSortByTargetRef(b *testing.B) {
	registry.RegisterTypeIfAbsent(v1alpha1.MeshHTTPRouteResourceTypeDescriptor)
	for _, n := range []int{10, 50, 200} {
		b.Run(fmt.Sprintf("policies=%d", n), func(b *testing.B) {
			list := &v1alpha1.MeshHTTPRouteResourceList{}
			for i := range n {
				list.Items = append(list.Items, &v1alpha1.MeshHTTPRouteResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "mesh-1",
						Name: fmt.Sprintf("route-%d", n-i),
						Labels: map[string]string{
							mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
							mesh_proto.PolicyRoleLabel:     string(mesh_proto.ConsumerPolicyRole),
							mesh_proto.DisplayName:         fmt.Sprintf("route-%d", n-i),
							mesh_proto.KubeNamespaceTag:    "ns-a",
						},
					},
					Spec: &v1alpha1.MeshHTTPRoute{
						TargetRef: &common_api.TopLevelTargetRef{
							Kind:   common_api.TopLevelTargetRefKindDataplane,
							Labels: pointer.To(map[string]string{"app": fmt.Sprintf("app-%d", i%10)}),
						},
					},
				})
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = matchers.SortByTargetRef(list)
			}
		})
	}
}
