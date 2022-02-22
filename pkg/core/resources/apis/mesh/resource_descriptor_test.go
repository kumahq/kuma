package mesh

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Descriptor", func() {
	It("should create new object from descriptor", func() {
		Expect(NewDataplaneResource().Descriptor().NewObject()).ToNot(BeNil())
	})

	It("should create new list from descriptor", func() {
		Expect(NewDataplaneResource().Descriptor().NewList()).ToNot(BeNil())
	})
})
