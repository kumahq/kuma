package model_test

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/kds"
	"github.com/kumahq/kuma/v2/pkg/kds/context"
	"github.com/kumahq/kuma/v2/pkg/kds/reconcile"
	policies_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
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
	metaFuncs := map[string]map[string]func(mesh, name string) model.ResourceMeta{
		"zone": {
			"k8s": func(mesh, name string) model.ResourceMeta {
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
			"universal": func(mesh, name string) model.ResourceMeta {
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
			"k8s": func(mesh, name string) model.ResourceMeta {
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
			"universal": func(mesh, name string) model.ResourceMeta {
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

	syncTo := func(meta func(mesh, name string) model.ResourceMeta, dst string) func(mesh, name string) model.ResourceMeta {
		return func(mesh, name string) model.ResourceMeta {
			gm := meta(mesh, name)

			var mapper reconcile.ResourceMapper
			switch dst {
			case "global":
				mapper = context.HashSuffixMapper(false, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag)
			case "zone":
				mapper = context.HashSuffixMapper(true)
			}

			r := v1alpha1.NewMeshTimeoutResource() // resource doesn't matter, we just want to call mapper to get a new meta
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
				Expect(model.IsReferenced(refMeta("m1", "t1"), "route-1", resMeta("m1", "route-1"))).To(BeTrue())
			})

			It("should return false when t1 is referencing route-2", func() {
				Expect(model.IsReferenced(refMeta("m1", "t1"), "route-2", resMeta("m1", "route-1"))).To(BeFalse())
			})

			It("should return false when meshes are different", func() {
				Expect(model.IsReferenced(refMeta("m1", "t1"), "route-1", resMeta("m2", "route-1"))).To(BeFalse())
			})

			It("should return true when route name has max allowed length", func() {
				longRouteName := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab"
				Expect(model.IsReferenced(refMeta("m1", "t1"), longRouteName, resMeta("m1", longRouteName))).To(BeTrue())
			})
		}, entries,
	)
})
