package events_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/events"
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
		eventBus := events.NewEventBus(1)
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
	})

	It("should only send events matched predicate", func() {
		// given
		eventBus := events.NewEventBus(10)
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
	})
})
