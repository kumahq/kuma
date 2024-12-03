package graph_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
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
			g := graph.BuildGraph(services, given.mtps)

			// then
			for from := range services {
				for to := range services {
					_, fromAll := given.expectedFromAll[to]
					_, conn := given.expectedConnections[from][to]
					Expect(g.CanReach(
						map[string]string{mesh_proto.ServiceTag: from},
						map[string]string{mesh_proto.ServiceTag: to},
					)).To(Equal(fromAll || conn))
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
		Entry("allow all no top targetRef", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
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
		Entry("equal subsets matching is preserved because of the names", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithName("bbb").
					WithTargetRef(builders.TargetRefMeshSubset("kuma.io/zone", "east")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Deny).
					Build(),
				builders.MeshTrafficPermission().
					WithName("aaa").
					WithTargetRef(builders.TargetRefMeshSubset("version", "v1")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			expectedFromAll: fromAllServices,
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
		graph := graph.BuildGraph(services, mtps)

		// then
		Expect(graph.CanReach(
			map[string]string{mesh_proto.ServiceTag: "b", "version": "v1"},
			map[string]string{mesh_proto.ServiceTag: "a"},
		)).To(BeTrue())
		Expect(graph.CanReach(
			map[string]string{mesh_proto.ServiceTag: "b", "version": "v2"},
			map[string]string{mesh_proto.ServiceTag: "a"},
		)).To(BeFalse())
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
		graph := graph.BuildGraph(services, mtps)

		// then
		Expect(graph.CanReach(
			map[string]string{"kuma.io/zone": "east"},
			map[string]string{mesh_proto.ServiceTag: "a"},
		)).To(BeTrue())
		Expect(graph.CanReach(
			map[string]string{"kuma.io/zone": "west"},
			map[string]string{mesh_proto.ServiceTag: "a"},
		)).To(BeFalse())
		Expect(graph.CanReach(
			map[string]string{"othertag": "other"},
			map[string]string{mesh_proto.ServiceTag: "a"},
		)).To(BeFalse())
	})

	It("should always allow cross mesh", func() {
		// when
		graph := graph.BuildGraph(nil, nil)

		// then
		Expect(graph.CanReach(
			map[string]string{mesh_proto.ServiceTag: "b"},
			map[string]string{mesh_proto.ServiceTag: "a", mesh_proto.MeshTag: "other"},
		)).To(BeTrue())
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
			graph := graph.BuildGraph(services, mtps)

			// then
			Expect(graph.CanReach(
				map[string]string{mesh_proto.ServiceTag: "b"},
				map[string]string{mesh_proto.ServiceTag: "a_kuma-demo_svc_1234"},
			)).To(BeTrue())
			Expect(graph.CanReach(
				map[string]string{mesh_proto.ServiceTag: "a_kuma-demo_svc_1234"},
				map[string]string{mesh_proto.ServiceTag: "b"},
			)).To(BeFalse()) // it's not selected by top-level target ref
		},
		Entry("MeshSubset by kube namespace", builders.TargetRefMeshSubset(mesh_proto.KubeNamespaceTag, "kuma-demo")),
		Entry("MeshSubset by kube service name", builders.TargetRefMeshSubset(mesh_proto.KubeServiceTag, "a")),
		Entry("MeshSubset by kube service port", builders.TargetRefMeshSubset(mesh_proto.KubePortTag, "1234")),
		Entry("MeshServiceSubset by kube namespace", builders.TargetRefServiceSubset("a_kuma-demo_svc_1234", mesh_proto.KubeNamespaceTag, "kuma-demo")),
		Entry("MeshServiceSubset by kube service name", builders.TargetRefServiceSubset("a_kuma-demo_svc_1234", mesh_proto.KubeServiceTag, "a")),
		Entry("MeshServiceSubset by kube service port", builders.TargetRefServiceSubset("a_kuma-demo_svc_1234", mesh_proto.KubePortTag, "1234")),
	)

	It("should not modify MeshTrafficPermission passed to the func when replacing tags in subsets", func() {
		// given
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

		// when
		_ = graph.BuildGraph(services, mtps)

		// then
		Expect(mtps[0].Spec.TargetRef.Tags).NotTo(BeNil())
		Expect(mtps[1].Spec.TargetRef.Tags).NotTo(BeNil())
	})

	It("should build services adding supported tags and including external services", func() {
		// given
		tags := map[string]string{
			mesh_proto.ServiceTag:       "a_kuma-demo_svc_1234",
			mesh_proto.KubeNamespaceTag: "kuma-demo",
			mesh_proto.KubeServiceTag:   "a",
			mesh_proto.KubePortTag:      "1234",
		}
		dpps := []*mesh.DataplaneResource{
			samples.DataplaneBackendBuilder().
				WithAddress("1.1.1.1").
				WithInboundOfTagsMap(tags).
				Build(),
			samples.DataplaneBackendBuilder().
				WithAddress("1.1.1.2").
				WithInboundOfTagsMap(tags).
				Build(),
			samples.DataplaneBackendBuilder().
				WithAddress("1.1.1.3").
				WithServices("b", "c").
				Build(),
		}
		es := []*mesh.ExternalServiceResource{
			{
				Spec: &mesh_proto.ExternalService{
					Tags: map[string]string{
						mesh_proto.ServiceTag: "es-1",
					},
				},
			},
		}
		zis := []*mesh.ZoneIngressResource{
			builders.ZoneIngress().
				AddSimpleAvailableService("d").
				AddSimpleAvailableService("e").
				Build(),
		}

		// when
		services := graph.BuildServices("default", dpps, es, zis)

		// then
		Expect(services).To(Equal(map[string]mesh_proto.SingleValueTagSet{
			"a_kuma-demo_svc_1234": map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.KubeServiceTag:   "a",
				mesh_proto.KubePortTag:      "1234",
			},
			"b":    map[string]string{},
			"c":    map[string]string{},
			"d":    map[string]string{},
			"e":    map[string]string{},
			"es-1": map[string]string{},
		}))
	})
})
