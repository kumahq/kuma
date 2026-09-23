package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshaccesslog_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/kds/samples"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("ComputePolicyRole", func() {
	type testCase struct {
		policy       core_model.Policy
		namespace    resource_labels.Namespace
		zone         string
		expectedRole mesh_proto.PolicyRole
		expectedErr  string
	}

	// a to[] item that names one MeshService the way a producer policy has to
	ownService := func(name, namespace, zone string) common_api.TargetRef {
		return builders.TargetRefMeshServiceLabels(map[string]string{
			mesh_proto.DisplayName:      name,
			mesh_proto.KubeNamespaceTag: namespace,
			mesh_proto.ZoneTag:          zone,
		}, "")
	}

	idleTimeout := meshtimeout_api.Conf{
		IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
	}

	policyWithTo := func(refs ...common_api.TargetRef) core_model.Policy {
		b := builders.MeshTimeout().
			WithMesh("mesh-1").WithName("name-1").
			WithTargetRef(builders.TargetRefMesh())
		for _, ref := range refs {
			b = b.AddTo(ref, idleTimeout)
		}
		return b.Build().Spec
	}

	DescribeTable("should compute the correct policy role",
		func(given testCase) {
			role, err := resource_labels.ComputePolicyRole(given.policy, given.namespace, given.zone)
			if given.expectedErr != "" {
				Expect(err.Error()).To(Equal(given.expectedErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(role).To(Equal(given.expectedRole))
		},
		Entry("consumer policy", testCase{
			policy:       policyWithTo(builders.TargetRefMesh()),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("producer policy", testCase{
			policy:       policyWithTo(ownService("backend", "kuma-demo", "zone-1")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		Entry("producer policy for MeshHTTPRoute", testCase{
			policy: policyWithTo(builders.TargetRefMeshHTTPRouteLabels(map[string]string{
				mesh_proto.DisplayName:      "route-1",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.ZoneTag:          "zone-1",
			})),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		// A producer policy is applied to every dataplane in the mesh and synced to
		// the other zones, so it may only name a resource its own namespace owns. A
		// selector that leaves out the namespace or the zone is a subset match over
		// the whole mesh and resolves to another tenant's resource. Every such item
		// is a consumer item and stays inside the policy's namespace.
		Entry("consumer policy when to[] omits the namespace and the zone", testCase{
			policy: policyWithTo(builders.TargetRefMeshServiceLabels(map[string]string{
				mesh_proto.DisplayName: "backend",
			}, "")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy when to[] omits the zone", testCase{
			policy: policyWithTo(builders.TargetRefMeshServiceLabels(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}, "")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy when to[] omits the namespace", testCase{
			policy: policyWithTo(builders.TargetRefMeshServiceLabels(map[string]string{
				mesh_proto.DisplayName: "backend",
				mesh_proto.ZoneTag:     "zone-1",
			}, "")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy when to[] names another namespace", testCase{
			policy:       policyWithTo(ownService("backend", "other-ns", "zone-1")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy when to[] names another zone", testCase{
			policy:       policyWithTo(ownService("backend", "kuma-demo", "zone-2")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy when to[] carries an extra selector label", testCase{
			policy: policyWithTo(builders.TargetRefMeshServiceLabels(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.ZoneTag:          "zone-1",
				"app":                       "backend",
			}, "")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy when the policy has no zone of its own", testCase{
			policy:       policyWithTo(ownService("backend", "kuma-demo", "zone-1")),
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "",
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("workload-owner policy with rules", testCase{
			policy: builders.MeshAccessLog().
				WithTargetRef(builders.TargetRefMesh()).
				AddRule(builders.MeshAccessLogConf().
					AddBackends(make([]meshaccesslog_api.Backend, 0))).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			zone:         "zone-1",
			expectedRole: mesh_proto.WorkloadOwnerPolicyRole,
		}),
		Entry("policy with consumer and producer to-items", testCase{
			policy: policyWithTo(
				ownService("backend-1", "kuma-demo", "zone-1"),
				ownService("backend-2", "other-ns", "zone-1"),
			),
			namespace:   resource_labels.NewNamespace("kuma-demo", false),
			zone:        "zone-1",
			expectedErr: "it's not allowed to mix producer and consumer items in the same policy",
		}),
	)
})

var _ = Describe("Compute", func() {
	legacyProxyTypeLabel := "kuma.io/" + "proxy-type"

	type testCase struct {
		r              core_model.Resource
		mode           core.CpMode
		isK8s          bool
		localZone      string
		expectedLabels map[string]string
	}

	DescribeTable("should return correct label map",
		func(given testCase) {
			labels, err := resource_labels.Compute(resource_labels.Write{
				Descriptor:  given.r.Descriptor(),
				Spec:        given.r.GetSpec(),
				Namespace:   resource_labels.GetNamespace(given.r.GetMeta(), "kuma-system"),
				Mesh:        given.r.GetMeta().GetMesh(),
				DisplayName: given.r.GetMeta().GetName(),
				Labels:      given.r.GetMeta().GetLabels(),
			}, resource_labels.ControlPlane{Mode: given.mode, IsK8s: given.isK8s, Zone: given.localZone})
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
				"kuma.io/display-name": "idle-timeout",
				"kuma.io/env":          "kubernetes",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
			},
		}),
		Entry("source/destination policy on zone-k8s", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: &meshexternalservice_api.MeshExternalServiceResource{
				Spec: builders.MeshExternalService().Build().Spec,
				Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "sample-timeout"},
			},
			expectedLabels: map[string]string{
				"kuma.io/display-name": "sample-timeout",
				"kuma.io/env":          "kubernetes",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
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
			expectedLabels: map[string]string{
				"kuma.io/display-name": "mesh-1",
				"kuma.io/origin":       "zone",
			},
		}),
		Entry("plugin originated policy on global", testCase{
			mode:  core.Global,
			isK8s: true,
			r: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("idle-timeout").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "idle-timeout",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "global",
			},
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
				"kuma.io/display-name":  "idle-timeout",
				"kuma.io/policy-role":   "consumer",
				"kuma.io/mesh":          "mesh-1",
				"kuma.io/origin":        "zone",
				"kuma.io/zone":          "zone-1",
				"kuma.io/env":           "kubernetes",
			},
		}),
		Entry("user-supplied k8s.kuma.io/namespace label is overwritten with the real namespace", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			// lives in app-ns, carries a label pointing at other-ns
			r: func() core_model.Resource {
				r := builders.MeshTimeout().
					WithMesh("mesh-1").
					WithName("idle-timeout").
					WithNamespace("app-ns").
					WithTargetRef(builders.TargetRefMesh()).
					AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
						IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
					}).
					Build()
				r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag] = "other-ns"
				return r
			}(),
			expectedLabels: map[string]string{
				"k8s.kuma.io/namespace": "app-ns",
				"kuma.io/display-name":  "idle-timeout",
				"kuma.io/policy-role":   "consumer",
				"kuma.io/mesh":          "mesh-1",
				"kuma.io/origin":        "zone",
				"kuma.io/zone":          "zone-1",
				"kuma.io/env":           "kubernetes",
			},
		}),
		Entry("spoofed env, namespace and service-account labels are cleaned on universal zone", testCase{
			mode:      core.Zone,
			isK8s:     false,
			localZone: "zone-1",
			r: func() core_model.Resource {
				r := builders.MeshTimeout().
					WithMesh("mesh-1").
					WithName("idle-timeout").
					WithTargetRef(builders.TargetRefMesh()).
					AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
						IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
					}).
					Build()
				r.GetMeta().GetLabels()[mesh_proto.EnvTag] = "kubernetes"
				r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag] = "victim"
				r.GetMeta().GetLabels()[metadata.KumaServiceAccount] = "admin"
				return r
			}(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "idle-timeout",
				"kuma.io/env":          "universal",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
			},
		}),
		Entry("spoofed env label is overwritten on k8s zone", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: func() core_model.Resource {
				r := builders.MeshTimeout().
					WithMesh("mesh-1").
					WithName("idle-timeout").
					WithTargetRef(builders.TargetRefMesh()).
					AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
						IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
					}).
					Build()
				r.GetMeta().GetLabels()[mesh_proto.EnvTag] = "universal"
				return r
			}(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "idle-timeout",
				"kuma.io/env":          "kubernetes",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
			},
		}),
		Entry("spoofed zone label is overwritten with the local zone", testCase{
			mode:      core.Zone,
			isK8s:     false,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithName("backend-1").
				WithServices("backend").
				WithMesh("mesh-1").
				WithLabels(map[string]string{mesh_proto.ZoneTag: "other-zone"}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "backend-1",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
				"kuma.io/env":          "universal",
			},
		}),
		Entry("zone label is omitted when the zone has no name", testCase{
			mode:  core.Zone,
			isK8s: false,
			r: builders.Dataplane().
				WithName("backend-1").
				WithServices("backend").
				WithMesh("mesh-1").
				WithLabels(map[string]string{mesh_proto.ZoneTag: "other-zone"}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "backend-1",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/env":          "universal",
			},
		}),
		Entry("namespace and service-account labels are kept on k8s zone", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: func() core_model.Resource {
				r := builders.MeshTimeout().
					WithMesh("mesh-1").
					WithName("idle-timeout").
					WithNamespace("app-ns").
					WithTargetRef(builders.TargetRefMesh()).
					AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
						IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
					}).
					Build()
				r.GetMeta().GetLabels()[metadata.KumaServiceAccount] = "sa-1"
				return r
			}(),
			expectedLabels: map[string]string{
				"k8s.kuma.io/namespace":       "app-ns",
				"k8s.kuma.io/service-account": "sa-1",
				"kuma.io/display-name":        "idle-timeout",
				"kuma.io/policy-role":         "consumer",
				"kuma.io/mesh":                "mesh-1",
				"kuma.io/origin":              "zone",
				"kuma.io/zone":                "zone-1",
				"kuma.io/env":                 "kubernetes",
			},
		}),
		Entry("dataplane proxy", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithName("backend-1").
				WithServices("backend").
				WithMesh("mesh-1").
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "backend-1",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
				"kuma.io/env":          "kubernetes",
			},
		}),
		Entry("dataplane with ZoneIngress listener", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithMesh("mesh-1").
				With(func(dp *mesh.DataplaneResource) {
					dp.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
						{
							Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
							Address: "127.0.0.1",
							Port:    10001,
						},
					}
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name":         "dp-1",
				"kuma.io/mesh":                 "mesh-1",
				"kuma.io/origin":               "zone",
				"kuma.io/zone":                 "zone-1",
				"kuma.io/env":                  "kubernetes",
				"kuma.io/listener-zoneingress": "enabled",
			},
		}),
		Entry("dataplane with ZoneEgress listener", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithMesh("mesh-1").
				With(func(dp *mesh.DataplaneResource) {
					dp.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
						{
							Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
							Address: "127.0.0.1",
							Port:    10001,
						},
					}
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name":        "dp-1",
				"kuma.io/mesh":                "mesh-1",
				"kuma.io/origin":              "zone",
				"kuma.io/zone":                "zone-1",
				"kuma.io/env":                 "kubernetes",
				"kuma.io/listener-zoneegress": "enabled",
			},
		}),
		Entry("dataplane with both ZoneIngress and ZoneEgress listeners", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithMesh("mesh-1").
				With(func(dp *mesh.DataplaneResource) {
					dp.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
						{
							Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
							Address: "127.0.0.1",
							Port:    10001,
						},
						{
							Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
							Address: "127.0.0.1",
							Port:    10002,
						},
					}
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name":         "dp-1",
				"kuma.io/mesh":                 "mesh-1",
				"kuma.io/origin":               "zone",
				"kuma.io/zone":                 "zone-1",
				"kuma.io/env":                  "kubernetes",
				"kuma.io/listener-zoneingress": "enabled",
				"kuma.io/listener-zoneegress":  "enabled",
			},
		}),
		Entry("stale listener labels are removed when listeners are absent", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithMesh("mesh-1").
				WithServices("backend").
				WithLabels(map[string]string{
					legacyProxyTypeLabel:           "sidecar",
					"kuma.io/listener-zoneingress": "enabled",
					"kuma.io/listener-zoneegress":  "enabled",
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/display-name": "dp-1",
				"kuma.io/mesh":         "mesh-1",
				"kuma.io/origin":       "zone",
				"kuma.io/zone":         "zone-1",
				"kuma.io/env":          "kubernetes",
			},
		}),
	)

	It("does not recompute labels for imported resources on privileged writes", func() {
		// A resource synced from another CP: on this zone it is not locally
		// originated (origin=global) and its display-name was set by the origin
		// CP. A privileged (KDS sync) write must leave those labels untouched.
		res := builders.MeshTimeout().
			WithMesh("mesh-1").
			WithName("idle-timeout").
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
				IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
			}).
			Build()
		existing := map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			mesh_proto.MeshTag:             "mesh-1",
			mesh_proto.DisplayName:         "name-from-origin-cp",
		}

		labels, err := resource_labels.Compute(resource_labels.Write{
			Descriptor:    res.Descriptor(),
			Spec:          res.GetSpec(),
			Mesh:          "mesh-1",
			DisplayName:   "recomputed-name",
			Labels:        existing,
			TrustedWriter: true,
		}, resource_labels.ControlPlane{Mode: core.Zone, IsK8s: true, Zone: "zone-1"})

		Expect(err).ToNot(HaveOccurred())
		Expect(labels).To(Equal(existing))
	})

	It("recomputes labels on privileged writes to locally-originated resources", func() {
		// A privileged write (origin=zone on this zone) is locally originated,
		// so it must still recompute CP-owned labels: the stale display-name is
		// overwritten, unlike the imported-resource case above.
		res := builders.MeshTimeout().
			WithMesh("mesh-1").
			WithName("idle-timeout").
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
				IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
			}).
			Build()
		existing := map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			mesh_proto.MeshTag:             "mesh-1",
			mesh_proto.DisplayName:         "stale-name",
		}

		labels, err := resource_labels.Compute(resource_labels.Write{
			Descriptor:    res.Descriptor(),
			Spec:          res.GetSpec(),
			Mesh:          "mesh-1",
			DisplayName:   "recomputed-name",
			Labels:        existing,
			TrustedWriter: true,
		}, resource_labels.ControlPlane{Mode: core.Zone, IsK8s: true, Zone: "zone-1"})

		Expect(err).ToNot(HaveOccurred())
		Expect(labels).To(HaveKeyWithValue(mesh_proto.DisplayName, "recomputed-name"))
	})

	DescribeTable("control-plane-only labels",
		func(trusted bool, expectKept bool) {
			res := builders.MeshService().WithMesh("mesh-1").WithName("backend").Build()
			supplied := map[string]string{
				mesh_proto.ResourceOriginLabel:             string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ManagedByLabel:                  "k8s-controller",
				mesh_proto.DeletionGracePeriodStartedLabel: "2026-01-01T00.00.00Z",
				metadata.KumaServiceName:                   "backend",
				metadata.HeadlessService:                   "false",
			}

			labels, err := resource_labels.Compute(resource_labels.Write{
				Descriptor:    res.Descriptor(),
				Spec:          res.GetSpec(),
				Namespace:     resource_labels.NewNamespace("kuma-demo", false),
				Mesh:          "mesh-1",
				DisplayName:   "backend",
				Labels:        supplied,
				TrustedWriter: trusted,
			}, resource_labels.ControlPlane{Mode: core.Zone, IsK8s: true, Zone: "zone-1"})

			Expect(err).ToNot(HaveOccurred())
			for _, key := range []string{mesh_proto.ManagedByLabel, mesh_proto.DeletionGracePeriodStartedLabel, metadata.KumaServiceName, metadata.HeadlessService} {
				if expectKept {
					Expect(labels).To(HaveKeyWithValue(key, supplied[key]))
				} else {
					Expect(labels).ToNot(HaveKey(key))
				}
			}
		},
		Entry("are kept as supplied by a trusted writer", true, true),
		Entry("are dropped when supplied by a user", false, false),
	)
})
