package watchdog_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

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

	It("should wait for the first tick to happen on WaitForFirstTick", func() {
		watchdog := &SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func(ctx context.Context) error {
				return nil
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}
		watchdog.WithTickCheck()

		// setup
		hasTicked := make(chan struct{})
		go func() {
			watchdog.HasTicked(true)
			close(hasTicked)
		}()
		go func() {
			watchdog.Start(ctx)
			close(doneCh)
		}()
		Expect(watchdog.HasTicked(false)).Should(BeFalse())
		Consistently(hasTicked).ShouldNot(Receive())

		By("simulating 1st tick")
		// when
		timeTicks <- time.Time{}
		Expect(watchdog.HasTicked(false)).Should(BeTrue())
		Expect(hasTicked).Should(BeClosed())

		By("simulating 2nd tick")
		// when
		timeTicks <- time.Time{}
		Expect(watchdog.HasTicked(false)).Should(BeTrue())
		Expect(hasTicked).Should(BeClosed())

		cancel()
		Eventually(doneCh).Should(BeClosed())
	})

	It("should not produce error on context cancelled", func() {
		// given
		watchdog := &SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return &time.Ticker{
					C: timeTicks,
				}
			},
			OnTick: func(ctx context.Context) error {
				return errors.New("missing data")
			},
			OnError: func(err error) {
				onErrorCalls <- err
			},
		}

		// when context is cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		watchdog.Start(ctx)

		// then no error is produced
		select {
		case <-onErrorCalls:
			Fail("error should not be produced")
		default:
		}
	})
})
