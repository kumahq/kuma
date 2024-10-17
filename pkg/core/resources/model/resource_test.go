package model_test

import (
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/context"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("Resource", func() {
	It("should return a new resource object", func() {
		// given
		desc := policies_api.MeshAccessLogResourceTypeDescriptor

		// when
		obj := desc.NewObject()

		// then
		Expect(reflect.TypeOf(obj.GetSpec()).String()).To(Equal("*v1alpha1.MeshAccessLog"))
		Expect(reflect.ValueOf(obj.GetSpec()).IsNil()).To(BeFalse())
	})
})

var _ = Describe("IsReferenced", func() {
	metaFuncs := map[string]map[string]func(mesh, name string) core_model.ResourceMeta{
		"zone": {
			"k8s": func(mesh, name string) core_model.ResourceMeta {
				return &test_model.ResourceMeta{
					Mesh: mesh,
					Name: fmt.Sprintf("%s.foo", name),
					Labels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
						mesh_proto.ZoneTag:             "zone-1",
						mesh_proto.KubeNamespaceTag:    "kuma-system",
						mesh_proto.DisplayName:         name,
					},
				}
			},
			"universal": func(mesh, name string) core_model.ResourceMeta {
				return &test_model.ResourceMeta{
					Mesh: mesh,
					Name: name,
					Labels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
						mesh_proto.ZoneTag:             "zone-1",
					},
				}
			},
		},
		"global": {
			"k8s": func(mesh, name string) core_model.ResourceMeta {
				return &test_model.ResourceMeta{
					Mesh: mesh,
					Name: fmt.Sprintf("%s.foo", name),
					Labels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
						mesh_proto.KubeNamespaceTag:    "kuma-system",
						mesh_proto.DisplayName:         name,
					},
				}
			},
			"universal": func(mesh, name string) core_model.ResourceMeta {
				return &test_model.ResourceMeta{
					Mesh: mesh,
					Name: name,
					Labels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				}
			},
		},
	}

	syncTo := func(meta func(mesh, name string) core_model.ResourceMeta, dst string) func(mesh, name string) core_model.ResourceMeta {
		return func(mesh, name string) core_model.ResourceMeta {
			gm := meta(mesh, name)

			var mapper reconcile_v2.ResourceMapper
			switch dst {
			case "global":
				mapper = context.HashSuffixMapper(false, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag)
			case "zone":
				mapper = context.HashSuffixMapper(true)
			}

			r := meshtimeout_api.NewMeshTimeoutResource() // resource doesn't matter, we just want to call mapper to get a new meta
			r.SetMeta(gm)
			nr, err := mapper(kds.Features{}, r)
			Expect(err).ToNot(HaveOccurred())
			return nr.GetMeta()
		}
	}

	type testCase struct {
		methodIsCalledOn string
		refOrigin        string
		resOrigin        string
		clusterTypes     map[string]string
	}

	var entries []TableEntry
	for _, methodIsCalledOn := range []string{"zone", "global"} {
		for _, refOrigin := range []string{"zone", "global"} {
			for _, resOrigin := range []string{"zone", "global"} {
				for _, zoneType := range []string{"k8s", "universal"} {
					for _, globalType := range []string{"k8s", "universal"} {
						description := fmt.Sprintf("on=%s, ref_origin=%s, res_origin=%s, zone_type=%s, global_type=%s", methodIsCalledOn, refOrigin, resOrigin, zoneType, globalType)
						entries = append(entries, Entry(description, testCase{
							methodIsCalledOn: methodIsCalledOn,
							refOrigin:        refOrigin,
							resOrigin:        resOrigin,
							clusterTypes: map[string]string{
								"zone":   zoneType,
								"global": globalType,
							},
						}))
					}
				}
			}
		}
	}

	DescribeTableSubtree("",
		func(g testCase) {
			refMeta := metaFuncs[g.refOrigin][g.clusterTypes[g.refOrigin]]
			resMeta := metaFuncs[g.resOrigin][g.clusterTypes[g.resOrigin]]

			if g.methodIsCalledOn != g.refOrigin {
				refMeta = syncTo(refMeta, g.methodIsCalledOn)
			}

			if g.methodIsCalledOn != g.resOrigin {
				resMeta = syncTo(resMeta, g.methodIsCalledOn)
			}

			It("should return true when t1 is referencing route-1", func() {
				Expect(core_model.IsReferenced(refMeta("m1", "t1"), "route-1", resMeta("m1", "route-1"))).To(BeTrue())
			})

			It("should return false when t1 is referencing route-2", func() {
				Expect(core_model.IsReferenced(refMeta("m1", "t1"), "route-2", resMeta("m1", "route-1"))).To(BeFalse())
			})

			It("should return false when meshes are different", func() {
				Expect(core_model.IsReferenced(refMeta("m1", "t1"), "route-1", resMeta("m2", "route-1"))).To(BeFalse())
			})

			It("should return true when route name has max allowed length", func() {
				longRouteName := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab"
				Expect(core_model.IsReferenced(refMeta("m1", "t1"), longRouteName, resMeta("m1", longRouteName))).To(BeTrue())
			})
		}, entries,
	)
})

var _ = Describe("ComputePolicyRole", func() {
	type testCase struct {
		policy       core_model.Policy
		namespace    core_model.Namespace
		expectedRole mesh_proto.PolicyRole
		expectedErr  string
	}

	DescribeTable("should compute the correct policy role",
		func(given testCase) {
			role, err := core_model.ComputePolicyRole(given.policy, given.namespace)
			if given.expectedErr != "" {
				Expect(err.Error()).To(Equal(given.expectedErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(role).To(Equal(given.expectedRole))
		},
		Entry("consumer policy", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    core_model.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy with labels", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
					"kuma.io/display-name":  "test",
					"kuma.io/zone":          "zone-1",
					"k8s.kuma.io/namespace": "kuma-demo",
				}, ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    core_model.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("producer policy", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "kuma-demo", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    core_model.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		Entry("producer policy with no namespace in to[]", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    core_model.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		Entry("workload-owner policy with from", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    core_model.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.WorkloadOwnerPolicyRole,
		}),
		Entry("workload-owner policy with both from and to", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:   core_model.NewNamespace("kuma-demo", false),
			expectedErr: "it's not allowed to mix 'to' and 'from' arrays in the same policy",
		}),
		Entry("consumer policy with from", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "backend-ns", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:   core_model.NewNamespace("kuma-demo", false),
			expectedErr: "it's not allowed to mix 'to' and 'from' arrays in the same policy",
		}),
		Entry("system policy with both from and to", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    core_model.NewNamespace("kuma-system", true),
			expectedRole: mesh_proto.SystemPolicyRole,
		}),
		Entry("policy with consumer and producer to-itmes", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend-1", "backend-1-ns", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddTo(builders.TargetRefMeshService("backend-2", "backend-2-ns", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:   core_model.NewNamespace("backend-1-ns", false),
			expectedErr: "it's not allowed to mix producer and consumer items in the same policy",
		}),
	)
})

var _ = Describe("ComputeLabels", func() {
	type testCase struct {
		r              core_model.Resource
		mode           core.CpMode
		isK8s          bool
		localZone      string
		expectedLabels map[string]string
	}

	DescribeTable("should return correct label map",
		func(given testCase) {
			labels, err := core_model.ComputeLabels(
				given.r.Descriptor(),
				given.r.GetSpec(),
				given.r.GetMeta().GetLabels(),
				core_model.GetNamespace(given.r.GetMeta(), "kuma-system"),
				given.r.GetMeta().GetMesh(),
				given.mode,
				given.isK8s,
				given.localZone,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(labels).To(Equal(given.expectedLabels))
		},
		Entry("plugin originated policy on zone-k8s", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("idle-timeout").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/env":    "kubernetes",
				"kuma.io/mesh":   "mesh-1",
				"kuma.io/origin": "zone",
				"kuma.io/zone":   "zone-1",
			},
		}),
		Entry("source/destination policy on zone-k8s", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: &mesh.TimeoutResource{
				Spec: samples.Timeout,
				Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "sample-timeout"},
			},
			expectedLabels: map[string]string{
				"kuma.io/mesh": "mesh-1",
			},
		}),
		Entry("mesh resource on non-federated zone", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: &mesh.MeshResource{
				Spec: samples.Mesh1,
				Meta: &test_model.ResourceMeta{Mesh: core_model.NoMesh, Name: "mesh-1"},
			},
			expectedLabels: map[string]string{},
		}),
		Entry("plugin originated policy on zone-k8s on custom namespace", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.MeshTimeout().
				WithMesh("mesh-1").
				WithName("idle-timeout").
				WithNamespace("custom-ns").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build(),
			expectedLabels: map[string]string{
				"k8s.kuma.io/namespace": "custom-ns",
				"kuma.io/policy-role":   "consumer",
				"kuma.io/mesh":          "mesh-1",
				"kuma.io/origin":        "zone",
				"kuma.io/zone":          "zone-1",
				"kuma.io/env":           "kubernetes",
			},
		}),
	)
})
