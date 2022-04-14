package watchdog_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
	. "github.com/kumahq/kuma/pkg/util/watchdog"
)

var _ = Describe("SimpleWatchdog", func() {

	var timeTicks chan time.Time
	var onTickCalls chan struct{}
	var onErrorCalls chan error

	BeforeEach(func() {
		timeTicks = make(chan time.Time)
		onTickCalls = make(chan struct{})
		onErrorCalls = make(chan error)
	})

	var stopCh, doneCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})
		doneCh = make(chan struct{})
	})

	It("should call OnTick() on timer ticks", test.Within(5*time.Second, func() {
		// given
		watchdog := SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func() error {
				onTickCalls <- struct{}{}
				return nil
			},
		}

		// setup
		go func() {
			watchdog.Start(stopCh)

			close(doneCh)
		}()

		By("simulating 1st tick")
		// when
		timeTicks <- time.Time{}

		// then
		<-onTickCalls

		By("simulating 2nd tick")
		// when
		timeTicks <- time.Time{}

		// then
		<-onTickCalls

		By("simulating Dataplane disconnect")
		// when
		close(stopCh)

		// then
		<-doneCh
	}))

	It("should call OnError() when OnTick() returns an error", test.Within(5*time.Second, func() {
		// given
		expectedErr := fmt.Errorf("expected error")
		// and
		watchdog := SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func() error {
				return expectedErr
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}

		// setup
		go func() {
			watchdog.Start(stopCh)

			close(doneCh)
		}()

		By("simulating 1st tick")
		// when
		timeTicks <- time.Time{}

		// then
		actualErr := <-onErrorCalls
		Expect(actualErr).To(MatchError(expectedErr))

		By("simulating Dataplane disconnect")
		// when
		close(stopCh)

		// then
		<-doneCh
	}))
})
