package naming_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/naming"
)

var _ = Describe("GetNameOrFallback", func() {
	It("should return name when predicate is true (bool)", func() {
		// when
		got := naming.GetNameOrFallback(true, "name", "fallback")
		// then
		Expect(got).To(Equal("name"))
	})

	It("should return fallback when predicate is false (bool)", func() {
		// when
		got := naming.GetNameOrFallback(false, "name", "fallback")
		// then
		Expect(got).To(Equal("fallback"))
	})

	It("should return name when predicate func returns true", func() {
		pred := func() bool { return true }
		got := naming.GetNameOrFallback(pred, "name", "fallback")
		Expect(got).To(Equal("name"))
	})

	It("should return fallback when predicate func returns false", func() {
		pred := func() bool { return false }
		got := naming.GetNameOrFallback(pred, "name", "fallback")
		Expect(got).To(Equal("fallback"))
	})

	It("should return fallback when predicate func is nil", func() {
		var pred func() bool // nil
		got := naming.GetNameOrFallback(pred, "name", "fallback")
		Expect(got).To(Equal("fallback"))
	})
})

var _ = Describe("GetNameOrFallbackFunc", func() {
	It("should return function that uses name when bool predicate true", func() {
		fn := naming.GetNameOrFallbackFunc(true)
		Expect(fn("name", "fallback")).To(Equal("name"))
	})

	It("should return function that uses fallback when bool predicate false", func() {
		fn := naming.GetNameOrFallbackFunc(false)
		Expect(fn("name", "fallback")).To(Equal("fallback"))
	})

	It("should return function that uses name when predicate func returns true", func() {
		pred := func() bool { return true }
		fn := naming.GetNameOrFallbackFunc(pred)
		Expect(fn("name", "fallback")).To(Equal("name"))
	})

	It("should return function that uses fallback when predicate func returns false", func() {
		pred := func() bool { return false }
		fn := naming.GetNameOrFallbackFunc(pred)
		Expect(fn("name", "fallback")).To(Equal("fallback"))
	})

	It("should evaluate predicate at call time (dynamic)", func() {
		val := false
		pred := func() bool { return val }
		fn := naming.GetNameOrFallbackFunc(pred)

		// initially false -> fallback
		Expect(fn("name", "fallback")).To(Equal("fallback"))

		// change underlying value -> should now return name
		val = true
		Expect(fn("name", "fallback")).To(Equal("name"))

		// change back -> fallback again
		val = false
		Expect(fn("name", "fallback")).To(Equal("fallback"))
	})

	It("should return fallback when predicate func is nil", func() {
		var pred func() bool // nil
		fn := naming.GetNameOrFallbackFunc(pred)
		Expect(fn("name", "fallback")).To(Equal("fallback"))
	})
})
