package issuer_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/xds/issuer"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	test_metrics "github.com/kumahq/kuma/v3/pkg/test/metrics"
)

var _ = Describe("Limiter", func() {
	var now time.Time
	var metrics core_metrics.Metrics
	var limiter *issuer.Limiter

	const backoff = 5 * time.Second

	backend := kri.Identifier{ResourceType: "Mesh", Name: "default", SectionName: "ca-1"}
	proxy := func(i int) model.ResourceKey {
		return model.ResourceKey{Mesh: "default", Name: fmt.Sprintf("dp-%d", i)}
	}

	BeforeEach(func() {
		now = time.Now()
		core.Now = func() time.Time { return now }

		m, err := core_metrics.NewMetrics("local")
		Expect(err).ToNot(HaveOccurred())
		metrics = m

		limiter, err = issuer.NewLimiter(issuer.Config{
			NewBackoff: func() retry.Backoff { return retry.NewConstant(backoff) },
			MinProxies: 3,
			Window:     time.Minute,
			Cooldown:   30 * time.Second,
		}, m)
		Expect(err).ToNot(HaveOccurred())
	})

	openGauge := func() float64 {
		return test_metrics.FindMetric(metrics, "cert_generation_circuit_open").GetGauge().GetValue()
	}

	It("backs off a single proxy without ever opening the backend circuit", func() {
		p := proxy(1)
		// first failure allowed, then backing off
		ok, _ := limiter.Allow(backend, p)
		Expect(ok).To(BeTrue())
		limiter.Record(backend, p, false)

		// many more failures from the SAME proxy - all suppressed, circuit stays closed
		for range 100 {
			ok, retryAfter := limiter.Allow(backend, p)
			Expect(ok).To(BeFalse())
			Expect(retryAfter).To(BeNumerically(">", 0))
		}
		Expect(openGauge()).To(Equal(0.0))

		// backoff elapses -> allowed again
		now = now.Add(backoff + time.Second)
		ok, _ = limiter.Allow(backend, p)
		Expect(ok).To(BeTrue())
	})

	It("opens the backend circuit once enough distinct proxies fail", func() {
		// MinProxies distinct proxies each fail once
		for i := range 3 {
			p := proxy(i)
			ok, _ := limiter.Allow(backend, p)
			Expect(ok).To(BeTrue())
			limiter.Record(backend, p, false)
		}

		// circuit is now open: a brand-new proxy is rejected without attempting
		ok, retryAfter := limiter.Allow(backend, proxy(999))
		Expect(ok).To(BeFalse())
		Expect(retryAfter).To(BeNumerically(">", 0))
		Expect(openGauge()).To(Equal(1.0))
	})

	It("half-opens after cooldown and closes on a successful probe", func() {
		for i := range 3 {
			p := proxy(i)
			limiter.Allow(backend, p)
			limiter.Record(backend, p, false)
		}
		Expect(openGauge()).To(Equal(1.0))

		// within cooldown: still open
		now = now.Add(10 * time.Second)
		ok, _ := limiter.Allow(backend, proxy(0))
		Expect(ok).To(BeFalse())

		// cooldown elapsed: a single probe is allowed
		now = now.Add(30 * time.Second)
		ok, _ = limiter.Allow(backend, proxy(0))
		Expect(ok).To(BeTrue())
		// concurrent second caller is rejected while the probe is in flight
		ok, _ = limiter.Allow(backend, proxy(1))
		Expect(ok).To(BeFalse())

		// probe succeeds -> circuit closes
		limiter.Record(backend, proxy(0), true)
		Expect(openGauge()).To(Equal(0.0))
		ok, _ = limiter.Allow(backend, proxy(1))
		Expect(ok).To(BeTrue())
	})

	It("reopens when the half-open probe fails", func() {
		for i := range 3 {
			p := proxy(i)
			limiter.Allow(backend, p)
			limiter.Record(backend, p, false)
		}
		now = now.Add(31 * time.Second)

		ok, _ := limiter.Allow(backend, proxy(0)) // probe
		Expect(ok).To(BeTrue())
		limiter.Record(backend, proxy(0), false) // probe fails
		Expect(openGauge()).To(Equal(1.0))

		// still open within the new cooldown
		now = now.Add(10 * time.Second)
		ok, _ = limiter.Allow(backend, proxy(0))
		Expect(ok).To(BeFalse())
	})

	It("Forget clears a proxy's backoff", func() {
		p := proxy(1)
		limiter.Allow(backend, p)
		limiter.Record(backend, p, false)
		ok, _ := limiter.Allow(backend, p)
		Expect(ok).To(BeFalse()) // backing off

		limiter.Forget(p)

		ok, _ = limiter.Allow(backend, p)
		Expect(ok).To(BeTrue()) // state dropped, allowed again
	})

	It("is a no-op when nil", func() {
		var nilLimiter *issuer.Limiter
		ok, retryAfter := nilLimiter.Allow(backend, proxy(0))
		Expect(ok).To(BeTrue())
		Expect(retryAfter).To(Equal(time.Duration(0)))
		nilLimiter.Record(backend, proxy(0), false) // must not panic
	})
})
