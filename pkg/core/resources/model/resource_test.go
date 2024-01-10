package model_test

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/plugins/common/k8s"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
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
	Context("Universal", func() {
		meta := func(mesh, name string) *test_model.ResourceMeta {
			return &test_model.ResourceMeta{Mesh: mesh, Name: name}
		}
		It("should return true when t1 is referencing route-1", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-1", meta("m1", "route-1"))).To(BeTrue())
		})
		It("should return false when t1 is referencing route-2", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-2", meta("m1", "route-1"))).To(BeFalse())
		})
		It("should return false when meshes are different", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-1", meta("m2", "route-1"))).To(BeFalse())
		})
	})
	Context("Kubernetes", func() {
		meta := func(mesh, name string) *test_model.ResourceMeta {
			return &test_model.ResourceMeta{
				Mesh:           mesh,
				Name:           fmt.Sprintf("%s.foo", name),
				NameExtensions: k8s.ResourceNameExtensions("foo", name),
			}
		}
		It("should return true when t1 is referencing route-1", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-1", meta("m1", "route-1"))).To(BeTrue())
		})
		It("should return false when t1 is referencing route-2", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-2", meta("m1", "route-1"))).To(BeFalse())
		})
		It("should return false when meshes are different", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-1", meta("m2", "route-1"))).To(BeFalse())
		})
	})
	Context("Kubernetes Zone", func() {
		meta := func(mesh, name string) *test_model.ResourceMeta {
			return &test_model.ResourceMeta{
				Mesh:           mesh,
				Name:           fmt.Sprintf("%s.foo", hash.HashedName(mesh, name)),
				NameExtensions: k8s.ResourceNameExtensions("foo", hash.HashedName(mesh, name)),
			}
		}
		It("should return true when t1 is referencing route-1", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-1", meta("m1", "route-1"))).To(BeTrue())
		})
		It("should return true when route name has max allowed length", func() {
			longRouteName := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab"
			Expect(model.IsReferenced(meta("m1", "t1"), longRouteName, meta("m1", longRouteName))).To(BeTrue())
		})
		It("should return false when t1 is referencing route-2", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-2", meta("m1", "route-1"))).To(BeFalse())
		})
		It("should return false when meshes are different", func() {
			Expect(model.IsReferenced(meta("m1", "t1"), "route-1", meta("m2", "route-1"))).To(BeFalse())
		})
	})
})
