package once_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
)

var _ = Describe("OnceCache", func() {
	var metrics core_metrics.Metrics
	var cache *once.Cache

	expiration := time.Millisecond * 200

	BeforeEach(func() {
		var err error
		metrics, err = core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		cache, err = once.New(expiration, "cache", metrics)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should cache Get() queries", func() {
		var count int32 = 0
		var val int32 = 1
		fn := once.RetrieverFunc(func(ctx context.Context, s string) (interface{}, error) {
			atomic.AddInt32(&count, 1)
			v := atomic.LoadInt32(&val)
			return v, nil
		})
		By("getting item for the first time")
		out, err := cache.GetOrRetrieve(context.Background(), "k1", fn)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.(int32)).To(Equal(int32(1)))
		Expect(atomic.LoadInt32(&count)).To(Equal(int32(1)))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit")).To(BeNil())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error")).To(BeNil())

		By("getting cached item")
		out, err = cache.GetOrRetrieve(context.Background(), "k1", fn)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.(int32)).To(Equal(int32(1)))
		Expect(atomic.LoadInt32(&count)).To(Equal(int32(1)))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit").Gauge.GetValue()).To(Equal(1.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error")).To(BeNil())

		By("updating Dataplane in store")
		atomic.StoreInt32(&val, 2)

		By("cached value hasn't changed")
		out, err = cache.GetOrRetrieve(context.Background(), "k1", fn)
		Expect(err).ToNot(HaveOccurred())
		Expect(atomic.LoadInt32(&count)).To(Equal(int32(1)))
		Expect(out.(int32)).To(Equal(int32(1)))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit").Gauge.GetValue()).To(Equal(2.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error")).To(BeNil())

		By("wait for invalidation")
		time.Sleep(expiration)

		By("get new value")
		out, err = cache.GetOrRetrieve(context.Background(), "k1", fn)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.(int32)).To(Equal(int32(2)))
		Expect(atomic.LoadInt32(&count)).To(Equal(int32(2)))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(2.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit").Gauge.GetValue()).To(Equal(2.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error")).To(BeNil())
	})

	It("should cache concurrent Get() requests", func() {
		var count int32 = 0
		var val int32 = 1
		fn := once.RetrieverFunc(func(ctx context.Context, s string) (interface{}, error) {
			atomic.AddInt32(&count, 1)
			v := atomic.LoadInt32(&val)
			return v, nil
		})
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				out, err := cache.GetOrRetrieve(context.Background(), "key", fn)
				Expect(err).ToNot(HaveOccurred())

				Expect(out.(int32)).To(Equal(atomic.LoadInt32(&val)))
				wg.Done()
			}()
		}
		wg.Wait()

		Expect(atomic.LoadInt32(&count)).To(Equal(int32(1)))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error")).To(BeNil())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		hitWaits := 0.0
		if hw := test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit-wait"); hw != nil {
			hitWaits = hw.Gauge.GetValue()
		}
		hits := 0.0
		if h := test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit"); h != nil {
			hits = h.Gauge.GetValue()
		}
		Expect(hitWaits + hits + 1).To(Equal(100.0))
	})

	It("should retry previously failed Get() requests", func() {
		var count int32 = 0
		var hasError int32 = 1
		fn := once.RetrieverFunc(func(ctx context.Context, s string) (interface{}, error) {
			atomic.AddInt32(&count, 1)
			if atomic.LoadInt32(&hasError) != 0 {
				return "", errors.New("it's an error")
			}
			return "hello", nil
		})
		By("getting Hash for the first time")
		out, err := cache.GetOrRetrieve(context.Background(), "key-1", fn)
		Expect(err).To(HaveOccurred())
		Expect(out).To(BeEmpty())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit")).To(BeNil())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error").Gauge.GetValue()).To(Equal(1.0))

		By("getting Hash calls again")
		out, err = cache.GetOrRetrieve(context.Background(), "key-1", fn)
		Expect(err).To(HaveOccurred())
		Expect(out).To(BeEmpty())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(2.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit")).To(BeNil())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "error").Gauge.GetValue()).To(Equal(2.0))

		By("Getting the hash once manager is fixed")
		atomic.StoreInt32(&hasError, 0)
		out, err = cache.GetOrRetrieve(context.Background(), "key-1", fn)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).ToNot(BeEmpty())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(3.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit")).To(BeNil())

		By("Now it should cache the hash once manager is fixed")
		out, err = cache.GetOrRetrieve(context.Background(), "key-1", fn)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(3.0))
		Expect(test_metrics.FindMetric(metrics, "cache", "operation", "get", "result", "hit").Gauge.GetValue()).To(Equal(1.0))
	})
})
