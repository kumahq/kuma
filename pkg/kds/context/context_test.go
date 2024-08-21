package context_test

import (
	stdcontext "context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/kds/util"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Context", func() {
	Describe("ZoneResourceMapper", func() {
		var rm manager.ResourceManager
		var mapper reconcile_v2.ResourceMapper

		type testCase struct {
			resource model.Resource
			expect   model.Resource
		}

		BeforeEach(func() {
			cfg := kuma_cp.DefaultConfig()
			cfg.Store.Type = config_store.KubernetesStore
			cfg.Multizone.Zone.Name = "zone"

			rm = manager.NewResourceManager(memory.NewStore())
			defaultContext := context.DefaultContext(stdcontext.Background(), rm, cfg)
			mapper = defaultContext.ZoneResourceMapper
		})

		DescribeTable("should zero generation field",
			func(given testCase) {
				// when
				out, _ := mapper(kds.Features{}, given.resource)

				// then
				Expect(out.GetMeta()).To(Equal(util.CloneResourceMeta(given.expect.GetMeta())))
				Expect(out.Descriptor()).To(Equal(given.expect.Descriptor()))
				Expect(out.GetSpec()).To(matchers.MatchProto(given.expect.GetSpec()))
			},
			Entry("should zero generation on DataplaneInsight", testCase{
				resource: &core_mesh.DataplaneInsightResource{
					Meta: &test_model.ResourceMeta{
						Name: "dpi-1",
					},
					Spec: &mesh_proto.DataplaneInsight{
						MTLS: &mesh_proto.DataplaneInsight_MTLS{
							IssuedBackend:     "test",
							SupportedBackends: []string{"one", "two"},
						},
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								Id:         "sub1",
								Generation: 10,
							},
							{
								Id:         "sub2",
								Generation: 15,
							},
						},
					},
				},
				expect: &core_mesh.DataplaneInsightResource{
					Meta: &test_model.ResourceMeta{
						Name: hash.HashedName("", "dpi-1", "zone"),
						Labels: map[string]string{
							"kuma.io/origin":       "zone",
							"kuma.io/zone":         "zone",
							"kuma.io/display-name": "dpi-1",
						},
					},
					Spec: &mesh_proto.DataplaneInsight{
						MTLS: &mesh_proto.DataplaneInsight_MTLS{
							IssuedBackend:     "test",
							SupportedBackends: []string{"one", "two"},
						},
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								Id:         "sub1",
								Generation: 0,
							},
							{
								Id:         "sub2",
								Generation: 0,
							},
						},
					},
				},
			}),
			Entry("should zero generation on ZoneIngressInsight", testCase{
				resource: &core_mesh.ZoneIngressInsightResource{
					Meta: &test_model.ResourceMeta{
						Name: "zii-1",
					},
					Spec: &mesh_proto.ZoneIngressInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ControlPlaneInstanceId: "ID1",
								Generation:             10,
							},
							{
								ControlPlaneInstanceId: "ID2",
								Generation:             15,
							},
						},
					},
				},
				expect: &core_mesh.ZoneIngressInsightResource{
					Meta: &test_model.ResourceMeta{
						Name: hash.HashedName("", "zii-1", "zone"),
						Labels: map[string]string{
							"kuma.io/display-name": "zii-1",
							"kuma.io/origin":       "zone",
							"kuma.io/zone":         "zone",
						},
					},
					Spec: &mesh_proto.ZoneIngressInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ControlPlaneInstanceId: "ID1",
								Generation:             0,
							},
							{
								ControlPlaneInstanceId: "ID2",
								Generation:             0,
							},
						},
					},
				},
			}),
			Entry("should zero generation on ZoneEgressInsight", testCase{
				resource: &core_mesh.ZoneEgressInsightResource{
					Meta: &test_model.ResourceMeta{
						Name: "zei-1",
					},
					Spec: &mesh_proto.ZoneEgressInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								Generation: 10,
							},
							{
								Generation: 15,
							},
						},
					},
				},
				expect: &core_mesh.ZoneEgressInsightResource{
					Meta: &test_model.ResourceMeta{
						Name: hash.HashedName("", "zei-1", "zone"),
						Labels: map[string]string{
							"kuma.io/zone":         "zone",
							"kuma.io/display-name": "zei-1",
							"kuma.io/origin":       "zone",
						},
					},
					Spec: &mesh_proto.ZoneEgressInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								Generation: 0,
							},
							{
								Generation: 0,
							},
						},
					},
				},
			}),
			Entry("should not change non-insight", testCase{
				resource: &core_mesh.CircuitBreakerResource{
					Meta: &test_model.ResourceMeta{
						Name: "cb-1",
					},
					Spec: &mesh_proto.CircuitBreaker{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"match1": "source",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"match2": "dest",
								},
							},
						},
						Conf: &mesh_proto.CircuitBreaker_Conf{
							SplitExternalAndLocalErrors: true,
						},
					},
				},
				expect: &core_mesh.CircuitBreakerResource{
					Meta: &test_model.ResourceMeta{
						Name: hash.HashedName("", "cb-1", "zone"),
						Labels: map[string]string{
							"kuma.io/display-name": "cb-1",
							"kuma.io/origin":       "zone",
							"kuma.io/zone":         "zone",
						},
					},
					Spec: &mesh_proto.CircuitBreaker{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"match1": "source",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"match2": "dest",
								},
							},
						},
						Conf: &mesh_proto.CircuitBreaker_Conf{
							SplitExternalAndLocalErrors: true,
						},
					},
				},
			}),
		)
	})
	Describe("GlobalProvidedFilter", func() {
		var rm manager.ResourceManager
		var predicate reconcile_v2.ResourceFilter

		clusterID := "cluster-id"
		configs := map[string]bool{
			config_manager.ClusterIdConfigKey: true,
		}

		type testCase struct {
			resource model.Resource
			expect   bool

			// zone ingresses and egresses
			zoneResource *core_system.ZoneResource
			zoneName     string
		}

		BeforeEach(func() {
			rm = manager.NewResourceManager(memory.NewStore())
			predicate = context.GlobalProvidedFilter(rm, configs)
		})

		It("should filter out configs if not in provided argument", func() {
			ctx := stdcontext.Background()
			// given
			config1 := &core_system.ConfigResource{
				Meta: &test_model.ResourceMeta{
					Name: config_manager.ClusterIdConfigKey,
				},
			}
			config2 := &core_system.ConfigResource{
				Meta: &test_model.ResourceMeta{
					Name: "config-which-should-be-filtered-out",
				},
			}

			// when
			ok := predicate(ctx, clusterID, kds.Features{}, config1)

			// then
			Expect(ok).To(BeTrue())

			// when
			ok = predicate(ctx, clusterID, kds.Features{}, config2)

			// then
			Expect(ok).To(BeFalse())
		})

		DescribeTable("global secrets",
			func(given testCase) {
				ctx := stdcontext.Background()
				// when
				ok := predicate(ctx, clusterID, kds.Features{
					kds.FeatureZoneToken: true,
				}, given.resource)

				// then
				Expect(ok).To(Equal(given.expect))
			},
			Entry("should not filter out zone token signing key", testCase{
				resource: &core_system.GlobalSecretResource{
					Meta: &test_model.ResourceMeta{
						Name: zone_tokens.SigningKeyPrefix + "-1",
					},
				},
				expect: true,
			}),
			Entry("should filter out when not signing key", testCase{
				resource: &core_system.GlobalSecretResource{
					Meta: &test_model.ResourceMeta{
						Name: "some-non-signing-key-global-secret",
					},
				},
				expect: false,
			}),
		)

		DescribeTable("zone ingresses",
			func(given testCase) {
				ctx := stdcontext.Background()
				// given
				if given.zoneResource != nil {
					Expect(rm.Create(
						ctx,
						given.zoneResource,
						core_store.CreateByKey(given.zoneName, ""),
					)).To(Succeed())
				}

				// when
				ok := predicate(ctx, clusterID, kds.Features{}, given.resource)

				// then
				Expect(ok).To(Equal(given.expect))
			},
			Entry("should not filter out zone ingresses from the different, enabled zone", testCase{
				resource: &core_mesh.ZoneIngressResource{
					Meta: &test_model.ResourceMeta{
						Name: "zone-ingress-1",
					},
					Spec: &mesh_proto.ZoneIngress{
						Zone: "different-zone",
					},
				},
				zoneResource: &core_system.ZoneResource{
					Meta: &test_model.ResourceMeta{
						Name: "different-zone",
					},
					Spec: &system_proto.Zone{
						Enabled: util_proto.Bool(true),
					},
				},
				zoneName: "different-zone",
				expect:   true,
			}),
			Entry("should filter out zone ingresses from the same zone", testCase{
				resource: &core_mesh.ZoneIngressResource{
					Meta: &test_model.ResourceMeta{
						Name: "zone-ingress-1",
					},
					Spec: &mesh_proto.ZoneIngress{
						Zone: clusterID,
					},
				},
				expect: false,
			}),
			Entry("should filter out zone ingresses from the different, not enabled zone", testCase{
				resource: &core_mesh.ZoneIngressResource{
					Meta: &test_model.ResourceMeta{
						Name: "zone-ingress-1",
					},
					Spec: &mesh_proto.ZoneIngress{
						Zone: "different-zone",
					},
				},
				zoneResource: &core_system.ZoneResource{
					Meta: &test_model.ResourceMeta{
						Name: "different-zone",
					},
					Spec: &system_proto.Zone{
						Enabled: util_proto.Bool(false),
					},
				},
				zoneName: "different-zone",
				expect:   false,
			}),
		)

		Context("global provided resources", func() {
			// we are ignoring this types, as we should already test them in
			// earlier tests
			ignoreTypes := map[model.ResourceType]struct{}{
				core_system.ConfigType:       {},
				core_system.GlobalSecretType: {},
				core_mesh.DataplaneType:      {},
				core_mesh.ZoneIngressType:    {},
				core_mesh.ZoneEgressType:     {},
			}

			var entries []TableEntry
			for _, descriptor := range registry.Global().ObjectDescriptors() {
				name := descriptor.Name
				_, ignoreType := ignoreTypes[name]

				if descriptor.KDSFlags.Has(model.GlobalToAllZonesFlag) && !ignoreType {
					resource := descriptor.NewObject()
					resource.SetMeta(&test_model.ResourceMeta{
						Name: string(name),
					})

					entries = append(entries, Entry(
						fmt.Sprintf("should return true for %s", name),
						testCase{resource: resource},
					))
				}
			}

			DescribeTable("returned predicate function",
				func(given testCase) {
					ctx := stdcontext.Background()
					// when
					ok := predicate(ctx, clusterID, kds.Features{}, given.resource)

					// then
					Expect(ok).To(BeTrue())
				},
				entries,
			)
		})
	})
	Describe("GlobalResourceMapper", func() {
		type config struct {
			storeType          config_store.StoreType
			k8sSystemNamespace string
		}

		type testCase struct {
			config                     config
			name                       string
			displayName                string
			expectedName               string
			isResourcePluginOriginated bool
			scope                      model.ResourceScope
			features                   kds.Features
			extraLabels                map[string]string
		}

		genConfig := func(caseCfg config) kuma_cp.Config {
			cfg := kuma_cp.DefaultConfig()

			if caseCfg.storeType != "" {
				cfg.Store.Type = caseCfg.storeType
			}

			if caseCfg.k8sSystemNamespace != "" {
				cfg.Store.Kubernetes.SystemNamespace = caseCfg.k8sSystemNamespace
			}

			return cfg
		}

		resource := func(given testCase) model.Resource {
			var r model.Resource = core_mesh.NewCircuitBreakerResource()
			switch given.scope {
			case model.ScopeGlobal:
				r = core_system.NewGlobalSecretResource()
			}
			if given.isResourcePluginOriginated {
				r = v1alpha1.NewMeshCircuitBreakerResource()
			}

			meta := &test_model.ResourceMeta{
				Name: given.name,
				Labels: map[string]string{
					mesh_proto.DisplayName: given.displayName,
				},
			}

			for k, v := range given.extraLabels {
				meta.GetLabels()[k] = v
			}

			r.SetMeta(meta)
			return r
		}

		DescribeTable("system namespace suffix from in resource names",
			func(given testCase) {
				// given
				ctx := stdcontext.Background()
				rm := manager.NewResourceManager(memory.NewStore())
				cfg := genConfig(given.config)
				kdsCtx := context.DefaultContext(ctx, rm, cfg)

				// when
				out, err := kdsCtx.GlobalResourceMapper(given.features, resource(given))
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(out.GetMeta().GetName()).To(Equal(given.expectedName))
			},
			Entry("should be removed when store type is kubernetes "+
				"and resource is plugin originated", testCase{
				isResourcePluginOriginated: true,
				config: config{
					storeType:          config_store.KubernetesStore,
					k8sSystemNamespace: "custom-namespace",
				},
				name:         "foo.custom-namespace",
				displayName:  "foo",
				expectedName: "foo-zxw6c95d42zfz9cc",
				scope:        model.ScopeMesh,
				features: map[string]bool{
					kds.FeatureHashSuffix: true,
				},
				extraLabels: map[string]string{},
			}),
			Entry("should be removed when store type is kubernetes "+
				"resource is plugin originated and no KDS hash-suffix feature", testCase{
				isResourcePluginOriginated: true,
				config: config{
					storeType:          config_store.KubernetesStore,
					k8sSystemNamespace: "custom-namespace",
				},
				name:         "foo.custom-namespace",
				expectedName: "foo",
				scope:        model.ScopeMesh,
				features:     map[string]bool{},
			}),
			Entry("shouldn't be removed when store type is kubernetes "+
				"and resource isn't plugin originated", testCase{
				isResourcePluginOriginated: false,
				config: config{
					storeType:          config_store.KubernetesStore,
					k8sSystemNamespace: "custom-namespace",
				},
				name:         "foo.default",
				expectedName: "foo.default",
				scope:        model.ScopeGlobal,
				features: map[string]bool{
					kds.FeatureHashSuffix: true,
				},
			}),
			Entry("should include zone and namespace in tags when they are present in labels", testCase{
				isResourcePluginOriginated: true,
				config: config{
					storeType:          config_store.KubernetesStore,
					k8sSystemNamespace: "custom-namespace",
				},
				name:         "foo.custom-namespace",
				displayName:  "foo",
				expectedName: "foo-696vzv497z4cv4f4",
				scope:        model.ScopeMesh,
				features: map[string]bool{
					kds.FeatureHashSuffix: true,
				},
				extraLabels: map[string]string{
					mesh_proto.ZoneTag:          "zone-1",
					mesh_proto.KubeNamespaceTag: "custom-ns",
				},
			}),
		)
	})
})
