package rules_test

import (
	"fmt"
	"maps"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/hash"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/k8s"
)

var _ = Describe("BuildResourceRules", func() {
	DescribeTableSubtree("BuildToRules",
		func(inputFile string) {
			type metaFn func(name, mesh string, labels map[string]string) core_model.ResourceMeta

			globalUni := metaFn(func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
				globalLabels := map[string]string{
					"kuma.io/origin":       "global",
					"kuma.io/display-name": name,
				}
				maps.Copy(globalLabels, labels)
				return &model.ResourceMeta{
					Name:   name,
					Mesh:   mesh,
					Labels: globalLabels,
				}
			})

			globalK8s := metaFn(func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
				globalLabels := map[string]string{
					"kuma.io/origin":        "global",
					"k8s.kuma.io/namespace": "ns-k8s",
					"kuma.io/mesh":          mesh,
					"kuma.io/display-name":  name,
				}
				maps.Copy(globalLabels, labels)
				return &model.ResourceMeta{
					Name:   k8s.K8sNamespacedNameToCoreName(name, "ns-k8s"),
					Mesh:   mesh,
					Labels: globalLabels,
					NameExtensions: map[string]string{
						"k8s.kuma.io/namespace": "ns-k8s",
						"k8s.kuma.io/name":      name,
					},
				}
			})

			zoneUni := metaFn(func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
				zoneLabels := map[string]string{
					"kuma.io/origin":       "zone",
					"kuma.io/zone":         "zone-uni",
					"kuma.io/display-name": name,
				}
				maps.Copy(zoneLabels, labels)
				return &model.ResourceMeta{
					Name:   name,
					Mesh:   mesh,
					Labels: zoneLabels,
				}
			})

			zoneK8s := metaFn(func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
				zoneLabels := map[string]string{
					"kuma.io/origin":        "zone",
					"kuma.io/zone":          "zone-k8s",
					"k8s.kuma.io/namespace": "ns-k8s",
					"kuma.io/mesh":          mesh,
					"kuma.io/display-name":  name,
				}
				maps.Copy(zoneLabels, labels)
				return &model.ResourceMeta{
					Name:   k8s.K8sNamespacedNameToCoreName(name, "ns-k8s"),
					Mesh:   mesh,
					Labels: zoneLabels,
					NameExtensions: map[string]string{
						"k8s.kuma.io/namespace": "ns-k8s",
						"k8s.kuma.io/name":      name,
					},
				}
			})

			syncToUni := func(fn metaFn) metaFn {
				return func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
					m := fn(name, mesh, labels)
					var values []string
					if v, ok := m.GetLabels()[mesh_proto.ZoneTag]; ok {
						values = append(values, v)
					}
					if v, ok := m.GetLabels()[mesh_proto.KubeNamespaceTag]; ok {
						values = append(values, v)
					}
					return &model.ResourceMeta{
						Name:   hash.HashedName(m.GetMesh(), core_model.GetDisplayName(m), values...),
						Mesh:   m.GetMesh(),
						Labels: m.GetLabels(),
					}
				}
			}

			syncToK8s := func(fn metaFn) metaFn {
				return func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
					m := fn(name, mesh, labels)
					var values []string
					if v, ok := m.GetLabels()[mesh_proto.ZoneTag]; ok {
						values = append(values, v)
					}
					if v, ok := m.GetLabels()[mesh_proto.KubeNamespaceTag]; ok {
						values = append(values, v)
					}
					newName := hash.HashedName(m.GetMesh(), core_model.GetDisplayName(m), values...)
					return &model.ResourceMeta{
						Name:   k8s.K8sNamespacedNameToCoreName(newName, "kuma-system"),
						Mesh:   m.GetMesh(),
						Labels: m.GetLabels(),
						NameExtensions: map[string]string{
							"k8s.kuma.io/namespace": "kuma-system",
							"k8s.kuma.io/name":      newName,
						},
					}
				}
			}

			updFn := func(fn metaFn, rs []core_model.Resource) {
				for _, r := range rs {
					if r.Descriptor().Name == mesh.MeshType {
						continue
					}
					r.SetMeta(fn(r.GetMeta().GetName(), r.GetMeta().GetMesh(), r.GetMeta().GetLabels()))
				}
			}

			type testCase struct {
				meta   metaFn
				golden string
			}
			DescribeTable("should build a rule-based view for policies",
				func(given testCase) {
					// given
					resources := readInputFile(inputFile)
					updFn(given.meta, resources)
					meshCtx := buildMeshContext(resources)
					toList, err := core_rules.BuildToList(matchedPolicies(resources), meshCtx)
					Expect(err).ToNot(HaveOccurred())

					// when
					rules, err := core_rules.BuildResourceRules(toList, meshCtx)
					Expect(err).ToNot(HaveOccurred())

					// then
					bytes, err := yaml.Marshal(rules)
					Expect(err).ToNot(HaveOccurred())
					Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", fmt.Sprintf(".%s.golden.", given.golden), 1)))
				},
				// policies are created and checked on the same cluster
				Entry("created and checked on global-uni", testCase{
					meta:   globalUni,
					golden: "global-uni",
				}),
				Entry("created and checked on global-k8s", testCase{
					meta:   globalK8s,
					golden: "global-k8s",
				}),
				Entry("created and checked on zone-uni", testCase{
					meta:   zoneUni,
					golden: "zone-uni",
				}),
				Entry("created and checked on zone-k8s", testCase{
					meta:   zoneK8s,
					golden: "zone-k8s",
				}),
				// policies are created on zone and checked on global
				Entry("created on zone-k8s, checked on global-uni", testCase{
					meta:   syncToUni(zoneK8s),
					golden: "zone-k8s-2-global-uni",
				}),
				Entry("created on zone-uni, checked on global-uni", testCase{
					meta:   syncToUni(zoneUni),
					golden: "zone-uni-2-global-uni",
				}),
				Entry("created on zone-k8s, checked on global-k8s", testCase{
					meta:   syncToK8s(zoneK8s),
					golden: "zone-k8s-2-global-k8s",
				}),
				Entry("created on zone-uni, checked on global-k8s", testCase{
					meta:   syncToK8s(zoneUni),
					golden: "zone-uni-2-global-k8s",
				}),
				// policies are created on global and checked on zone
				Entry("created on global-k8s, checked on zone-uni", testCase{
					meta:   syncToUni(globalK8s),
					golden: "global-k8s-2-zone-uni",
				}),
				Entry("created on global-uni, checked on zone-uni", testCase{
					meta:   syncToUni(globalUni),
					golden: "global-uni-2-zone-uni",
				}),
				Entry("created on global-uni, checked on zone-k8s", testCase{
					meta:   syncToK8s(globalUni),
					golden: "global-uni-2-zone-k8s",
				}),
				Entry("created on global-k8s, checked on zone-k8s", testCase{
					meta:   syncToK8s(globalK8s),
					golden: "global-k8s-2-zone-k8s",
				}),
				//
				Entry("created on zone-uni, checked on another zone-uni", testCase{
					meta:   syncToUni(syncToUni(zoneUni)),
					golden: "zone-uni-2-zone-uni",
				}),
				Entry("created on zone-uni, checked on another zone-k8s", testCase{
					meta:   syncToK8s(syncToUni(zoneUni)),
					golden: "zone-uni-2-zone-k8s",
				}),
				Entry("created on zone-k8s, checked on another zone-uni", testCase{
					meta:   syncToUni(syncToUni(zoneK8s)),
					golden: "zone-k8s-2-zone-uni",
				}),
				Entry("created on zone-k8s, checked on another zone-k8s", testCase{
					meta:   syncToK8s(syncToUni(zoneK8s)),
					golden: "zone-k8s-2-zone-k8s",
				}),
			)
		},
		test.EntriesForFolder("resourcerules"),
	)
})
