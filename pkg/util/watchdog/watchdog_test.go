package watchdog_test

import (
	"context"
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
	var doneCh chan struct{}
	var cancel context.CancelFunc
	var ctx context.Context

	BeforeEach(func() {
		timeTicks = make(chan time.Time)
		onTickCalls = make(chan struct{})
		onErrorCalls = make(chan error)
		doneCh = make(chan struct{})
		ctx, cancel = context.WithCancel(context.Background())
	})

	It("should call OnTick() on timer ticks", test.Within(5*time.Second, func() {
		// given
		watchdog := SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func(context.Context) error {
				onTickCalls <- struct{}{}
				return nil
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}

		// setup
		go func() {
			watchdog.Start(ctx)

			close(doneCh)
		}()

		// then first tick happens "immediately"
		By("ticks on first call")
		Eventually(onTickCalls).Should(Receive())

		By("simulating 1st tick")
		// when
		timeTicks <- time.Time{}

		// then
		Eventually(onTickCalls).Should(Receive())

		By("simulating 2nd tick")
		// when
		timeTicks <- time.Time{}

		// then
		<-onTickCalls

		By("simulating Dataplane disconnect")
		// when
		cancel()

		// then
		Eventually(doneCh).Should(BeClosed())
		Consistently(onErrorCalls).ShouldNot(Receive())
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
			OnTick: func(context.Context) error {
				return expectedErr
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}

		// setup
		go func() {
			watchdog.Start(ctx)

			close(doneCh)
		}()

		// then first tick happens "immediately"
		By("tick on startup")
		Eventually(onErrorCalls).Should(Receive(MatchError(expectedErr)))

		By("simulating 1st tick")
		// when
		timeTicks <- time.Time{}
		Eventually(onErrorCalls).Should(Receive(MatchError(expectedErr)))

		By("simulating Dataplane disconnect")
		// when
		cancel()

		// then
		Eventually(doneCh).Should(BeClosed())
	}))

	It("should not crash the whole application when watchdog crashes", test.Within(5*time.Second, func() {
		// given
		watchdog := SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func(context.Context) error {
				panic("xyz")
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}

		// when
		go func() {
			watchdog.Start(ctx)
			close(doneCh)
		}()

		// then watchdog returned an error
		Eventually(onErrorCalls).Should(Receive())
		cancel()
		Eventually(doneCh).Should(BeClosed())
	}))

	It("should cancel in flight tick when stopping", test.Within(5*time.Second, func() {
		// given
		watchdog := SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func(ctx context.Context) error {
				<-ctx.Done()
				onTickCalls <- struct{}{}
				return ctx.Err()
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}

		// setup
		go func() {
			watchdog.Start(ctx)

			close(doneCh)
		}()

		// then first tick is delayed because context is not cancelled
		By("ticks on first call")
		Consistently(onTickCalls).ShouldNot(Receive())

		By("simulating Dataplane disconnect")
		// when
		cancel()
		Eventually(onTickCalls).Should(Receive())

		// then
		Eventually(doneCh).Should(BeClosed())
		Consistently(onErrorCalls).ShouldNot(Receive())
	}))

	It("should skip ticks when StreamCtx is closed", test.Within(5*time.Second, func() {
		// given
		streamCtx, streamCancel := context.WithCancel(context.Background())
		onStopCalled := make(chan struct{})

		watchdog := SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func(context.Context) error {
				onTickCalls <- struct{}{}
				return nil
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
			OnStop: func(_ context.Context) {
				close(onStopCalled)
			},
			StreamCtx: streamCtx,
		}

		// setup
		go func() {
			watchdog.Start(ctx)
			close(doneCh)
		}()

		// then first tick happens "immediately"
		By("ticks on first call")
		Eventually(onTickCalls).Should(Receive())

		By("simulating 1st tick")
		timeTicks <- time.Time{}
		Eventually(onTickCalls).Should(Receive())

		By("simulating stream closure (gRPC disconnect)")
		streamCancel()

		By("simulating ticker fires after stream closed")
		timeTicks <- time.Time{}

		// then no tick should happen because stream is closed
		Consistently(onTickCalls, "100ms").ShouldNot(Receive())

		By("canceling main context to stop watchdog")
		cancel()

		// then watchdog should stop and call OnStop
		Eventually(onStopCalled).Should(BeClosed())
		Eventually(doneCh).Should(BeClosed())
	}))
})
