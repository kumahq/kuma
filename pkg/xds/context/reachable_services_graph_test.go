package context_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	v1alpha12 "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("Reachable Services Graph", func() {

	type testCase struct {
		mtps                []*v1alpha1.MeshTrafficPermissionResource
		expectedFromAll     map[string]struct{}
		expectedConnections map[string]map[string]struct{}
		expected            xds_context.ReachableServicesGraph
		assertion           func(xds_context.ReachableServicesGraph)
	}

	services := map[string]v1alpha12.SingleValueTagSet{
		"a": map[string]string{},
		"b": map[string]string{},
		"c": map[string]string{},
		"d": map[string]string{},
	}

	fromAllServices := map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}}

	DescribeTable("should generate graph",
		func(given testCase) {
			g, err := xds_context.BuildReachableServicesGraph(services, given.mtps)
			Expect(err).ToNot(HaveOccurred())

			for from := range services {
				for to := range services {
					_, fromAll := given.expectedFromAll[to]
					_, conn := given.expectedConnections[from][to]
					Expect(g.CanReach(map[string]string{v1alpha12.ServiceTag: from}, to)).To(Equal(fromAll || conn))
				}
			}
		},
		Entry("allow all", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: fromAllServices,
		}),
		Entry("deny all", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Deny).
					Build(),
			},
			expectedFromAll: map[string]struct{}{},
		}),
		Entry("no MeshTrafficPermissions", testCase{
			mtps:            []*v1alpha1.MeshTrafficPermissionResource{},
			expectedFromAll: map[string]struct{}{},
		}),
		Entry("one connection Allow", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFromX(builders.TargetRefService("a"), v1alpha1.Allow).
					Build(),
			},
			expectedConnections: map[string]map[string]struct{}{
				"a": {"b": {}},
			},
		}),
		Entry("AllowWithShadowDeny is treated as Allow", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFromX(builders.TargetRefService("a"), v1alpha1.AllowWithShadowDeny).
					Build(),
			},
			expectedConnections: map[string]map[string]struct{}{
				"a": {"b": {}},
			},
		}),
		Entry("multiple allowed connections", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFromX(builders.TargetRefService("a"), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("c")).
					AddFromX(builders.TargetRefService("b"), v1alpha1.AllowWithShadowDeny).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("d")).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: map[string]struct{}{
				"d": {},
			},
			expectedConnections: map[string]map[string]struct{}{
				"a": {"b": {}},
				"b": {"c": {}},
			},
		}),
		Entry("all allowed except one connection", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFromX(builders.TargetRefService("a"), v1alpha1.Deny).
					Build(),
			},
			expectedFromAll: map[string]struct{}{
				"a": {},
				"c": {},
				"d": {},
			},
			expectedConnections: map[string]map[string]struct{}{
				"c": {"b": {}},
				"d": {"b": {}},
				"b": {"b": {}},
			},
		}),
		Entry("allow all but one service has restrictive mesh traffic permission", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Deny).
					AddFromX(builders.TargetRefService("a"), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: map[string]struct{}{
				"a": {},
				"c": {},
				"d": {},
			},
			expectedConnections: map[string]map[string]struct{}{
				"a": {"b": {}},
			},
		}),

		Entry("top level target ref MeshSubset selects all", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMeshSubset("kuma.io/zone", "east")).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: fromAllServices,
		}),
		Entry("top level target ref MeshServiceSubset selects all instances of the service", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefServiceSubset("a", "kuma.io/zone", "east")).
					AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: map[string]struct{}{
				"a": {},
			},
		}),

		//Entry("allow only subset of the service", testCase{
		//	mtps: []*v1alpha1.MeshTrafficPermissionResource{
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefMesh()).
		//			AddFromX(builders.TargetRefMesh(), v1alpha1.Deny).
		//			Build(),
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefService("b")).
		//			AddFromX(builders.TargetRefServiceSubset("a", "version", "v1"), v1alpha1.Allow).
		//			AddFromX(builders.TargetRefServiceSubset("a", "version", "v2"), v1alpha1.Deny).
		//			Build(),
		//	},
		//	expected: xds_context.ReachableServicesGraph{
		//		FromAll: map[string]struct{}{},
		//		Connections: map[string]map[string]struct{}{
		//			"a": {"b": {}},
		//		},
		//	},
		//}),
		//Entry("allow only subset of the service and deny the rest", testCase{
		//	mtps: []*v1alpha1.MeshTrafficPermissionResource{
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefMesh()).
		//			AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
		//			Build(),
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefService("b")).
		//			AddFromX(builders.TargetRefService("a"), v1alpha1.Deny).
		//			AddFromX(builders.TargetRefServiceSubset("a", "version", "v1"), v1alpha1.Allow).
		//			Build(),
		//	},
		//	expected: fromAllToAllGraph,
		//}),
		//Entry("allow only subset of the service", testCase{
		//	mtps: []*v1alpha1.MeshTrafficPermissionResource{
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefMesh()).
		//			AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
		//			Build(),
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefService("b")).
		//			AddFromX(builders.TargetRefService("a"), v1alpha1.Allow).
		//			AddFromX(builders.TargetRefServiceSubset("a", "version", "v1"), v1alpha1.Deny).
		//			Build(),
		//	},
		//	expected: fromAllToAllGraph,
		//}),
		//Entry("allow mesh subset allows all", testCase{
		//	mtps: []*v1alpha1.MeshTrafficPermissionResource{
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefMesh()).
		//			AddFromX(builders.TargetRefMesh(), v1alpha1.Deny).
		//			AddFromX(builders.TargetRefMeshSubset("kuma.io/zone", "east"), v1alpha1.Allow).
		//			Build(),
		//	},
		//	expected: fromAllToAllGraph,
		//}),
		//Entry("deny mesh subset is ignored", testCase{
		//	mtps: []*v1alpha1.MeshTrafficPermissionResource{
		//		builders.MeshTrafficPermission().
		//			WithTargetRef(builders.TargetRefMesh()).
		//			AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
		//			AddFromX(builders.TargetRefMeshSubset("kuma.io/zone", "east"), v1alpha1.Deny).
		//			Build(),
		//	},
		//	expected: fromAllToAllGraph,
		//}),
	)

	It("should not modify MeshTrafficPermission when replacing subsets", func() {
		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMeshSubset("version", "v1")).
				AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
				Build(),
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefServiceSubset("a", "version", "v1")).
				AddFromX(builders.TargetRefMesh(), v1alpha1.Allow).
				Build(),
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFromX(builders.TargetRefMeshSubset("version", "v1"), v1alpha1.Allow).
				Build(),
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFromX(builders.TargetRefServiceSubset("a", "version", "v1"), v1alpha1.Allow).
				Build(),
		}

		_, err := xds_context.BuildReachableServicesGraph(services, mtps)

		Expect(err).ToNot(HaveOccurred())
		Expect(mtps[0].Spec.TargetRef.Kind).To(Equal(common_api.MeshSubset))
		Expect(mtps[0].Spec.TargetRef.Tags).NotTo(BeNil())
		Expect(mtps[1].Spec.TargetRef.Kind).To(Equal(common_api.MeshServiceSubset))
		Expect(mtps[1].Spec.TargetRef.Tags).NotTo(BeNil())
		Expect(mtps[2].Spec.From[0].TargetRef.Kind).To(Equal(common_api.MeshSubset))
		Expect(mtps[2].Spec.From[0].TargetRef.Tags).NotTo(BeNil())
		Expect(mtps[3].Spec.From[0].TargetRef.Kind).To(Equal(common_api.MeshServiceSubset))
		Expect(mtps[3].Spec.From[0].TargetRef.Tags).NotTo(BeNil())
	})
})
