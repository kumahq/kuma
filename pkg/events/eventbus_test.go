package events_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/events"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
)

var _ = Describe("EventBus", func() {
	chHadEvent := func(ch <-chan events.Event) bool {
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}

	It("should not block on Send", func() {
		// given
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		eventBus, err := events.NewEventBus(1, metrics)
		Expect(err).ToNot(HaveOccurred())
		listener := eventBus.Subscribe()
		event1 := events.ResourceChangedEvent{TenantID: "1"}
		event2 := events.ResourceChangedEvent{TenantID: "2"}

		// when
		eventBus.Send(event1)
		eventBus.Send(event2)

		// then
		event := <-listener.Recv()
		Expect(event).To(Equal(event1))

		// and second event was ignored because buffer was full
		Expect(chHadEvent(listener.Recv())).To(BeFalse())
		Expect(test_metrics.FindMetric(metrics, "events_dropped").Counter.GetValue()).To(Equal(1.0))
	})

	It("should only send events matched predicate", func() {
		// given
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		eventBus, err := events.NewEventBus(10, metrics)
		Expect(err).ToNot(HaveOccurred())
		listener := eventBus.Subscribe(func(event events.Event) bool {
			return event.(events.ResourceChangedEvent).TenantID == "1"
		})
		event1 := events.ResourceChangedEvent{TenantID: "1"}
		event2 := events.ResourceChangedEvent{TenantID: "2"}

		// when
		eventBus.Send(event1)
		eventBus.Send(event2)

		// then
		event := <-listener.Recv()
		Expect(event).To(Equal(event1))

		// and second event was ignored, because it did not match predicate
		Expect(chHadEvent(listener.Recv())).To(BeFalse())
		Expect(test_metrics.FindMetric(metrics, "events_dropped").Counter.GetValue()).To(Equal(0.0))
	})
})
