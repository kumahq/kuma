package model_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
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
