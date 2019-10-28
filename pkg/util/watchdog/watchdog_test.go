package watchdog_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/util/watchdog"
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

	It("should call OnTick() on timer ticks", func(done Done) {
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

		close(done)
	}, 5)

	It("should call OnError() when OnTick() returns an error", func(done Done) {
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

		close(done)
	}, 5)
})
