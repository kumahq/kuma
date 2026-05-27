package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("errSampler", func() {
	It("allows each key once per window", func() {
		sampler := newErrSampler()

		ok, suppressed := sampler.allow("demo/dp-1")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(0))

		ok, _ = sampler.allow("demo/dp-1")
		Expect(ok).To(BeFalse())

		ok, suppressed = sampler.allow("demo/dp-2")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(0))
	})

	It("counts suppressed calls", func() {
		sampler := newErrSampler()

		ok, _ := sampler.allow("demo/dp-1")
		Expect(ok).To(BeTrue())

		_, _ = sampler.allow("demo/dp-1")
		_, _ = sampler.allow("demo/dp-1")

		// expire the window
		sampler.last["demo/dp-1"] = time.Now().Add(-2 * errSamplerWindow)

		ok, suppressed := sampler.allow("demo/dp-1")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(2))
	})

	It("re-allows after window expires", func() {
		sampler := newErrSampler()
		sampler.last["demo/dp-1"] = time.Now().Add(-2 * errSamplerWindow)

		ok, suppressed := sampler.allow("demo/dp-1")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(0))
	})
})
