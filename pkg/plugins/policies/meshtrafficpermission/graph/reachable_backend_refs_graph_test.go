package graph_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph/backends"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

var _ = Describe("Reachable Backends Graph", func() {
	type testCase struct {
		mtps                []*v1alpha1.MeshTrafficPermissionResource
		expectedFromAll     map[string]struct{}
		expectedConnections map[string]map[string]struct{}
	}

	services := []*meshservice_api.MeshServiceResource{
		builders.MeshService().
			WithName("a").
			WithDataplaneTagsSelectorKV("app", "a", "k8s.kuma.io/namespace", "kuma-demo").
			Build(),
		builders.MeshService().
			WithName("b").
			WithDataplaneTagsSelectorKV("app", "b", "k8s.kuma.io/namespace", "kuma-demo").
			Build(),
		builders.MeshService().
			WithName("c").
			WithDataplaneTagsSelectorKV("app", "c", "k8s.kuma.io/namespace", "test").
			Build(),
		builders.MeshService().
			WithName("d").
			WithDataplaneTagsSelectorKV("app", "d", "k8s.kuma.io/namespace", "prod", "version", "v1").
			Build(),
	}

	fromAllServices := map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}}

	DescribeTable("should check reachability of the graph",
		func(given testCase) {
			// when
			g := graph.NewGraph(
				map[string]core_rules.Rules{},
				backends.BuildRules(services, given.mtps),
			)

			// then
			for _, from := range services {
				for _, to := range services {
					_, fromAll := given.expectedFromAll[to.Meta.GetName()]
					_, conn := given.expectedConnections[from.Meta.GetName()][to.Meta.GetName()]
					Expect(g.CanReachBackend(
						map[string]string{"app": from.Meta.GetName()},
						&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
							Kind: "MeshService",
							Name: to.Meta.GetName(),
						},
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
					WithTargetRef(builders.TargetRefMeshSubset("app", "b")).
					AddFrom(builders.TargetRefMeshSubset("app", "a"), v1alpha1.Allow).
					Build(),
			},
			expectedConnections: map[string]map[string]struct{}{
				"a": {"b": {}},
			},
		}),
		Entry("AllowWithShadowDeny is treated as Allow", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMeshSubset("app", "b")).
					AddFrom(builders.TargetRefMeshSubset("app", "a"), v1alpha1.AllowWithShadowDeny).
					Build(),
			},
			expectedConnections: map[string]map[string]struct{}{
				"a": {"b": {}},
			},
		}),
		Entry("multiple allowed connections", testCase{
			mtps: []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMeshSubset("app", "b")).
					AddFrom(builders.TargetRefMeshSubset("app", "a"), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMeshSubset("app", "c")).
					AddFrom(builders.TargetRefMeshSubset("app", "b"), v1alpha1.AllowWithShadowDeny).
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMeshSubset("app", "d")).
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
					WithTargetRef(builders.TargetRefMeshSubset("app", "b")).
					AddFrom(builders.TargetRefMeshSubset("app", "a"), v1alpha1.Deny).
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
					WithTargetRef(builders.TargetRefMeshSubset("app", "b")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Deny).
					AddFrom(builders.TargetRefMeshSubset("app", "a"), v1alpha1.Allow).
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
					WithTargetRef(builders.TargetRefMeshSubset("app", "a", "kuma.io/zone", "east")).
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
		services := []*meshservice_api.MeshServiceResource{
			builders.MeshService().
				WithName("a").
				WithDataplaneTagsSelectorKV("app", "a", "k8s.kuma.io/namespace", "kuma-demo").
				Build(),
		}
		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMeshSubset("app", "b", "version", "v1"), v1alpha1.Allow).
				Build(),
		}

		// when
		g := graph.NewGraph(
			map[string]core_rules.Rules{},
			backends.BuildRules(services, mtps),
		)
		// then
		Expect(g.CanReachBackend(
			map[string]string{"app": "b", "version": "v1"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "a",
			},
		)).To(BeTrue())
		Expect(g.CanReachBackend(
			map[string]string{mesh_proto.ServiceTag: "b", "version": "v2"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "a",
			},
		)).To(BeFalse())
	})

	It("should work with mesh subset in from", func() {
		services := []*meshservice_api.MeshServiceResource{
			builders.MeshService().
				WithName("a").
				WithDataplaneTagsSelectorKV("app", "a", "k8s.kuma.io/namespace", "kuma-demo").
				Build(),
		}
		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMeshSubset("kuma.io/zone", "east"), v1alpha1.Allow).
				Build(),
		}

		// when
		g := graph.NewGraph(
			map[string]core_rules.Rules{},
			backends.BuildRules(services, mtps),
		)

		// then
		Expect(g.CanReachBackend(
			map[string]string{"kuma.io/zone": "east"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "a",
			},
		)).To(BeTrue())
		Expect(g.CanReachBackend(
			map[string]string{"kuma.io/zone": "west"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "a",
			},
		)).To(BeFalse())
		Expect(g.CanReachBackend(
			map[string]string{"othertag": "other"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "a",
			},
		)).To(BeFalse())
	})

	DescribeTable("top level subset should work with predefined tags",
		func(targetRef common_api.TargetRef) {
			// given
			services := []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("a").
					WithDataplaneTagsSelectorKV("app", "a", "k8s.kuma.io/namespace", "kuma-demo", "k8s.kuma.io/service-name", "a", "k8s.kuma.io/service-port", "1234").
					Build(),
				builders.MeshService().
					WithName("b").
					WithDataplaneTagsSelectorKV("app", "b", "k8s.kuma.io/namespace", "not-matching", "k8s.kuma.io/service-name", "b", "k8s.kuma.io/service-port", "9999").
					Build(),
			}
			mtps := []*v1alpha1.MeshTrafficPermissionResource{
				builders.MeshTrafficPermission().
					WithTargetRef(targetRef).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			}

			// when
			g := graph.NewGraph(
				map[string]core_rules.Rules{},
				backends.BuildRules(services, mtps),
			)

			// then
			Expect(g.CanReachBackend(
				map[string]string{"app": "b"},
				&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
					Kind: "MeshService",
					Name: "a",
				},
			)).To(BeTrue())
			Expect(g.CanReachBackend(
				map[string]string{"app": "a"},
				&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
					Kind: "MeshService",
					Name: "b",
				},
			)).To(BeFalse()) // it's not selected by top-level target ref
		},
		Entry("MeshSubset by kube namespace", builders.TargetRefMeshSubset(mesh_proto.KubeNamespaceTag, "kuma-demo")),
		Entry("MeshSubset by kube service name", builders.TargetRefMeshSubset(mesh_proto.KubeServiceTag, "a")),
		Entry("MeshSubset by kube service port", builders.TargetRefMeshSubset(mesh_proto.KubePortTag, "1234")),
		Entry("MeshSubset by app and kube namespace", builders.TargetRefMeshSubset("app", "a", mesh_proto.KubeNamespaceTag, "kuma-demo")),
		Entry("MeshSubset by app and kube service name", builders.TargetRefMeshSubset("app", "a", mesh_proto.KubeServiceTag, "a")),
		Entry("MeshSubset by app and kube service port", builders.TargetRefMeshSubset("app", "a", mesh_proto.KubePortTag, "1234")),
	)

	It("should match only dp with specific labels", func() {
		// given
		services := []*meshservice_api.MeshServiceResource{
			builders.MeshService().
				WithName("a").
				WithDataplaneTagsSelectorKV("app", "a", "k8s.kuma.io/namespace", "kuma-demo", "k8s.kuma.io/service-name", "a", "k8s.kuma.io/service-port", "1234").
				Build(),
			builders.MeshService().
				WithName("b").
				WithDataplaneTagsSelectorKV("app", "b", "k8s.kuma.io/namespace", "not-available").
				Build(),
		}

		mtps := []*v1alpha1.MeshTrafficPermissionResource{
			builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMeshSubset(mesh_proto.KubeNamespaceTag, "kuma-demo")).
				AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
				Build(),
		}

		// when
		g := graph.NewGraph(
			map[string]core_rules.Rules{},
			backends.BuildRules(services, mtps),
		)

		// then
		Expect(g.CanReachBackend(
			map[string]string{"app": "b"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "a",
			},
		)).To(BeTrue())
		Expect(g.CanReachBackend(
			map[string]string{"app": "a"},
			&mesh_proto.Dataplane_Networking_Outbound_BackendRef{
				Kind: "MeshService",
				Name: "b",
			},
		)).To(BeFalse())
	})

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
		_ = backends.BuildRules(services, mtps)

		// then
		Expect(mtps[0].Spec.TargetRef.Tags).NotTo(BeNil())
		Expect(mtps[1].Spec.TargetRef.Tags).NotTo(BeNil())
	})
})
