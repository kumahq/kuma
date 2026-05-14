package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("errSampler", func() {
	It("allows each key once per window", func() {
		sampler := newErrSampler(time.Minute)

		Expect(sampler.allow("demo/dp-1:broken spec")).To(BeTrue())
		Expect(sampler.allow("demo/dp-1:broken spec")).To(BeFalse())
		Expect(sampler.allow("demo/dp-2:broken spec")).To(BeTrue())
	})

	It("prunes expired keys", func() {
		sampler := newErrSampler(time.Minute)
		sampler.last["demo/stale:broken spec"] = time.Now().Add(-2 * time.Minute)
		sampler.last["demo/fresh:broken spec"] = time.Now()

		Expect(sampler.allow("demo/new:broken spec")).To(BeTrue())
		Expect(sampler.last).NotTo(HaveKey("demo/stale:broken spec"))
		Expect(sampler.last).To(HaveKey("demo/fresh:broken spec"))
	})
})
