package context_test

import (
	stdcontext "context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

		It("should filter out dataplanes", func() {
			// given
			dp := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
				},
			}

			// when
			ok := predicate(clusterID, dp)

			// then
			Expect(ok).To(BeFalse())
		})

		It("should filter out configs if not in provided argument", func() {
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
			ok := predicate(clusterID, config1)

			// then
			Expect(ok).To(BeTrue())

			// when
			ok = predicate(clusterID, config2)

			// then
			Expect(ok).To(BeFalse())
		})

		DescribeTable("global secrets",
			func(given testCase) {
				// when
				ok := predicate(clusterID, given.resource)

				// then
				Expect(ok).To(Equal(given.expect))
			},
			Entry("should not filter out zone ingress token signing key", testCase{
				resource: &core_system.GlobalSecretResource{
					Meta: &test_model.ResourceMeta{
						Name: zoneingress.ZoneIngressSigningKeyPrefix + "-1",
					},
				},
				expect: true,
			}),
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
				// given
				if given.zoneResource != nil {
					Expect(rm.Create(
						stdcontext.Background(),
						given.zoneResource,
						core_store.CreateByKey(given.zoneName, ""),
					)).To(Succeed())
				}

				// when
				ok := predicate(clusterID, given.resource)

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

				if descriptor.KDSFlags.Has(model.ProvidedByGlobal) && !ignoreType {
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
					// when
					ok := predicate(clusterID, given.resource)

					// then
					Expect(ok).To(BeTrue())
				},
				entries...,
			)
		})
	})
})
