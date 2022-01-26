package context_test

import (
	stdcontext "context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Context", func() {
	Describe("GlobalProvidedFilter", func() {
		var rm manager.ResourceManager
		var predicate reconcile.ResourceFilter

		clusterID := "cluster-id"
		configs := map[string]bool{
			config_manager.ClusterIdConfigKey: true,
		}

		BeforeEach(func() {
			rm = manager.NewResourceManager(memory.NewStore())
			predicate = context.GlobalProvidedFilter(rm, configs)
		})

		It("Should filter out dataplanes", func() {
			// when
			dp := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
				},
			}

			// then
			Expect(predicate(clusterID, dp)).To(BeFalse())
		})

		It("Should filter out configs if not in provided argument", func() {
			// when
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

			// then
			Expect(predicate(clusterID, config1)).To(BeTrue())
			Expect(predicate(clusterID, config2)).To(BeFalse())
		})

		It("Should filter out global secrets if they are not signing keys", func() {
			// when
			// zone ingress signing key
			secret1 := &core_system.GlobalSecretResource{
				Meta: &test_model.ResourceMeta{
					Name: zoneingress.ZoneIngressSigningKeyPrefix + "-1",
				},
			}
			// zone token signing key
			secret2 := &core_system.GlobalSecretResource{
				Meta: &test_model.ResourceMeta{
					Name: zone_tokens.SigningKeyPrefix + "-1",
				},
			}
			// zone token signing key
			secret3 := &core_system.GlobalSecretResource{
				Meta: &test_model.ResourceMeta{
					Name: "some-other-global-secret",
				},
			}

			// then
			Expect(predicate(clusterID, secret1)).To(BeTrue())
			Expect(predicate(clusterID, secret2)).To(BeTrue())
			Expect(predicate(clusterID, secret3)).To(BeFalse())
		})

		It("Should filter out zone ingresses from the same zone", func() {
			// when
			// the same zone
			zoneIngress1 := &core_mesh.ZoneIngressResource{
				Meta: &test_model.ResourceMeta{
					Name: "zone-ingress-1",
				},
				Spec: &mesh_proto.ZoneIngress{
					Zone: clusterID,
				},
			}
			// different zones
			differentZone1Name := "zone-1"
			differentZone2Name := "zone-2"
			zoneIngress2 := &core_mesh.ZoneIngressResource{
				Meta: &test_model.ResourceMeta{
					Name: "zone-ingress-2",
				},
				Spec: &mesh_proto.ZoneIngress{
					Zone: differentZone1Name,
				},
			}
			zoneIngress3 := &core_mesh.ZoneIngressResource{
				Meta: &test_model.ResourceMeta{
					Name: "zone-ingress-3",
				},
				Spec: &mesh_proto.ZoneIngress{
					Zone: differentZone2Name,
				},
			}

			Expect(rm.Create(stdcontext.Background(), &core_system.ZoneResource{
				Meta: &test_model.ResourceMeta{
					Name: differentZone1Name,
				},
				Spec: &system_proto.Zone{
					Enabled: util_proto.Bool(true),
				},
			}, core_store.CreateByKey(differentZone1Name, ""))).To(Succeed())

			Expect(rm.Create(stdcontext.Background(), &core_system.ZoneResource{
				Meta: &test_model.ResourceMeta{
					Name: differentZone2Name,
				},
				Spec: &system_proto.Zone{
					Enabled: util_proto.Bool(false),
				},
			}, core_store.CreateByKey(differentZone2Name, ""))).To(Succeed())

			// then
			Expect(predicate(clusterID, zoneIngress1)).To(BeFalse())
			Expect(predicate(clusterID, zoneIngress2)).To(BeTrue())
			Expect(predicate(clusterID, zoneIngress3)).To(BeFalse())
		})

		It("Should filter out zone egresses from the same zone", func() {
			// when
			// the same zone
			zoneEgress1 := &core_mesh.ZoneEgressResource{
				Meta: &test_model.ResourceMeta{
					Name: "zone-egress-1",
				},
				Spec: &mesh_proto.ZoneEgress{
					Zone: clusterID,
				},
			}
			// different zones
			differentZone1Name := "zone-1"
			differentZone2Name := "zone-2"
			zoneEgress2 := &core_mesh.ZoneEgressResource{
				Meta: &test_model.ResourceMeta{
					Name: "zone-egress-2",
				},
				Spec: &mesh_proto.ZoneEgress{
					Zone: differentZone1Name,
				},
			}
			zoneEgress3 := &core_mesh.ZoneEgressResource{
				Meta: &test_model.ResourceMeta{
					Name: "zone-egress-3",
				},
				Spec: &mesh_proto.ZoneEgress{
					Zone: differentZone2Name,
				},
			}

			Expect(rm.Create(stdcontext.Background(), &core_system.ZoneResource{
				Meta: &test_model.ResourceMeta{
					Name: differentZone1Name,
				},
				Spec: &system_proto.Zone{
					Enabled: util_proto.Bool(true),
				},
			}, core_store.CreateByKey(differentZone1Name, ""))).To(Succeed())

			Expect(rm.Create(stdcontext.Background(), &core_system.ZoneResource{
				Meta: &test_model.ResourceMeta{
					Name: differentZone2Name,
				},
				Spec: &system_proto.Zone{
					Enabled: util_proto.Bool(false),
				},
			}, core_store.CreateByKey(differentZone2Name, ""))).To(Succeed())

			// then
			Expect(predicate(clusterID, zoneEgress1)).To(BeFalse())
			Expect(predicate(clusterID, zoneEgress2)).To(BeTrue())
			Expect(predicate(clusterID, zoneEgress3)).To(BeFalse())
		})

		It("Should filter resources with KDS Flags equal model.ProvidedByGlobal", func() {
			// when
			// we are ignoring this types, as we should already test them in
			// earlier tests
			ignoreTypes := map[model.ResourceType]struct{}{
				core_system.ConfigType:       {},
				core_system.GlobalSecretType: {},
				core_mesh.DataplaneType:      {},
				core_mesh.ZoneIngressType:    {},
				core_mesh.ZoneEgressType:     {},
			}

			// then
			for _, descriptor := range registry.Global().ObjectDescriptors() {
				_, ignoreType := ignoreTypes[descriptor.Name]

				if descriptor.KDSFlags.Has(model.ProvidedByGlobal) && !ignoreType {
					resource := descriptor.NewObject()
					resource.SetMeta(&test_model.ResourceMeta{
						Name: string(descriptor.Name),
					})

					Expect(predicate(clusterID, resource)).To(BeTrue())
				}
			}
		})
	})
})
