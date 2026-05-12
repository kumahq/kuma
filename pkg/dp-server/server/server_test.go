package server

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeHTTP struct {
	shutdownDuration time.Duration
}

func (f *fakeHTTP) Shutdown(ctx context.Context) error {
	select {
	case <-time.After(f.shutdownDuration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type fakeGRPC struct {
	stopCalled atomic.Bool
}

func (f *fakeGRPC) Stop() { f.stopCalled.Store(true) }

var _ = Describe("DpServer shutdown", func() {
	It("should stop gRPC after a clean drain", func() {
		h := &fakeHTTP{shutdownDuration: 10 * time.Millisecond}
		g := &fakeGRPC{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		Expect(shutdownDpServer(ctx, h, g)).To(Succeed())
		Expect(g.stopCalled.Load()).To(BeTrue())
	})

	It("should stop gRPC after the graceful shutdown deadline", func() {
		h := &fakeHTTP{shutdownDuration: 5 * time.Second}
		g := &fakeGRPC{}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := shutdownDpServer(ctx, h, g)
		Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
		Expect(g.stopCalled.Load()).To(BeTrue())
	})

	It("should bound the shutdown wall clock time", func() {
		h := &fakeHTTP{shutdownDuration: time.Hour}
		g := &fakeGRPC{}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		_ = shutdownDpServer(ctx, h, g)
		Expect(time.Since(start)).To(BeNumerically("<=", 200*time.Millisecond))
	})

	It("should block WaitForDone until the started server finishes", func() {
		d := &DpServer{done: make(chan struct{})}
		d.started.Store(true)
		waitDone := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			d.WaitForDone()
			close(waitDone)
		}()

		Consistently(waitDone, 10*time.Millisecond).ShouldNot(BeClosed())

		close(d.done)

		Eventually(waitDone).Should(BeClosed())
	})

	It("should return from WaitForDone when the server never started", func() {
		d := &DpServer{done: make(chan struct{})}
		waitDone := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			d.WaitForDone()
			close(waitDone)
		}()

		Eventually(waitDone).Should(BeClosed())
	})

	It("should refuse a second Start call", func() {
		d := &DpServer{done: make(chan struct{})}
		d.started.Store(true)

		err := d.Start(make(chan struct{}))
		Expect(err).To(MatchError("dp-server already started"))
	})
})
