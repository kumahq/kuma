package context_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("Reachable Services Graph", func() {
	type testCase struct {
		mtps                []*v1alpha1.MeshTrafficPermissionResource
		expectedFromAll     map[string]struct{}
		expectedConnections map[string]map[string]struct{}
	}

	services := map[string]mesh_proto.SingleValueTagSet{
		"a": map[string]string{},
		"b": map[string]string{},
		"c": map[string]string{},
		"d": map[string]string{},
	}

	fromAllServices := map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}}

	DescribeTable("should check reachability of the graph",
		func(given testCase) {
			// when
			g, err := xds_context.BuildReachableServicesGraph(services, given.mtps)

			// then
			Expect(err).ToNot(HaveOccurred())
			for from := range services {
				for to := range services {
					_, fromAll := given.expectedFromAll[to]
					_, conn := given.expectedConnections[from][to]
					Expect(g.CanReach(map[string]string{mesh_proto.ServiceTag: from}, to)).To(Equal(fromAll || conn))
				}
			}
		},
		Entry("allow all", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: fromAllServices,
		}),
		Entry("deny all", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Deny).
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
					AddFrom(builders.TargetRefService("a"), v1alpha1.Allow).
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
					AddFrom(builders.TargetRefService("a"), v1alpha1.AllowWithShadowDeny).
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
					AddFrom(builders.TargetRefService("a"), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("c")).
					AddFrom(builders.TargetRefService("b"), v1alpha1.AllowWithShadowDeny).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("d")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
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
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFrom(builders.TargetRefService("a"), v1alpha1.Deny).
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
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("b")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Deny).
					AddFrom(builders.TargetRefService("a"), v1alpha1.Allow).
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
		Entry("top level target ref MeshSubset with unsupported predefined tags selects all", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMeshSubset("kuma.io/zone", "east")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: fromAllServices,
		}),
		Entry("top level target ref MeshServiceSubset of unsupported predefined tags selects all instances of the service", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefServiceSubset("a", "kuma.io/zone", "east")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: map[string]struct{}{
				"a": {},
			},
		}),
	)

	It("should work with service subsets in from", func() {
		// given
		services := map[string]mesh_proto.SingleValueTagSet{
			"a": map[string]string{},
		}
		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefServiceSubset("b", "version", "v1"), v1alpha1.Allow).
				Build(),
		}

		// when
		graph, err := xds_context.BuildReachableServicesGraph(services, mtps)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(graph.CanReach(map[string]string{
			mesh_proto.ServiceTag: "b",
			"version":             "v1",
		}, "a")).To(BeTrue())
		Expect(graph.CanReach(map[string]string{
			mesh_proto.ServiceTag: "b",
			"version":             "v2",
		}, "a")).To(BeFalse())
	})

	It("should work with mesh subset in from", func() {
		services := map[string]mesh_proto.SingleValueTagSet{
			"a": map[string]string{},
		}
		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMeshSubset("kuma.io/zone", "east"), v1alpha1.Allow).
				Build(),
		}

		// when
		graph, err := xds_context.BuildReachableServicesGraph(services, mtps)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(graph.CanReach(map[string]string{"kuma.io/zone": "east"}, "a")).To(BeTrue())
		Expect(graph.CanReach(map[string]string{"kuma.io/zone": "west"}, "a")).To(BeFalse())
		Expect(graph.CanReach(map[string]string{"othertag": "other"}, "a")).To(BeFalse())
	})

	DescribeTable("top level subset should work with predefined tags",
		func(targetRef common_api.TargetRef) {
			// given
			services := map[string]mesh_proto.SingleValueTagSet{
				"a_kuma-demo_svc_1234": map[string]string{
					mesh_proto.KubeNamespaceTag: "kuma-demo",
					mesh_proto.KubeServiceTag:   "a",
					mesh_proto.KubePortTag:      "1234",
				},
				"b": map[string]string{},
			}
			mtps := []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(targetRef).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			}

			// when
			graph, err := xds_context.BuildReachableServicesGraph(services, mtps)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(graph.CanReach(map[string]string{mesh_proto.ServiceTag: "b"}, "a_kuma-demo_svc_1234")).To(BeTrue())
			Expect(graph.CanReach(map[string]string{mesh_proto.ServiceTag: "a_kuma-demo_svc_1234"}, "b")).To(BeFalse()) // it's not selected by top-level target ref
		},
		Entry("MeshSubset by kube namespace", builders.TargetRefMeshSubset(mesh_proto.KubeNamespaceTag, "kuma-demo")),
		Entry("MeshSubset by kube service name", builders.TargetRefMeshSubset(mesh_proto.KubeServiceTag, "a")),
		Entry("MeshSubset by kube service port", builders.TargetRefMeshSubset(mesh_proto.KubePortTag, "1234")),
		Entry("MeshServiceSubset by kube namespace", builders.TargetRefServiceSubset("a_kuma-demo_svc_1234", mesh_proto.KubeNamespaceTag, "kuma-demo")),
		Entry("MeshServiceSubset by kube service name", builders.TargetRefServiceSubset("a_kuma-demo_svc_1234", mesh_proto.KubeServiceTag, "a")),
		Entry("MeshServiceSubset by kube service port", builders.TargetRefServiceSubset("a_kuma-demo_svc_1234", mesh_proto.KubePortTag, "1234")),
	)

	It("should not modify MeshTrafficPermission passed to the func when replacing subsets", func() {
		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMeshSubset("version", "v1")).
				AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
				Build(),
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefServiceSubset("a", "version", "v1")).
				AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
				Build(),
		}

		_, err := xds_context.BuildReachableServicesGraph(services, mtps)

		Expect(err).ToNot(HaveOccurred())
		Expect(mtps[0].Spec.TargetRef.Kind).To(Equal(common_api.MeshSubset))
		Expect(mtps[0].Spec.TargetRef.Tags).NotTo(BeNil())
		Expect(mtps[1].Spec.TargetRef.Kind).To(Equal(common_api.MeshServiceSubset))
		Expect(mtps[1].Spec.TargetRef.Tags).NotTo(BeNil())
	})
})
