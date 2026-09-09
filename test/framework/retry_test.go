package framework

import (
	"context"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("retryKeepingLastError", func() {
	It("keeps the error from the last attempt", func() {
		// Given an action that always fails with a message naming the cause
		attempts := 0
		action := func() error {
			attempts++
			return errors.New("k3d image import: no space left on device")
		}

		// When the retries run out
		err := retryKeepingLastError(context.Background(), NewTestingT(), "k3d image import", 2, time.Millisecond, action)

		// Then the cause survives alongside terratest's own summary
		Expect(attempts).To(Equal(3)) // terratest counts retries after the first attempt
		Expect(err).To(MatchError(ContainSubstring("no space left on device")))
		Expect(err).To(MatchError(ContainSubstring("unsuccessful after 2 retries")))
	})

	It("keeps the cause rather than the kill that followed it", func() {
		// Given a budget that expires mid-run, and a first failure that names the cause.
		// The later attempts run a real command so the budget kills them the way it
		// kills a docker save - `signal: killed`, which is not a context error at all.
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		attempts := 0
		action := func() error {
			attempts++
			if attempts == 1 {
				return errors.New("ctr: content digest sha256:abc: not found")
			}
			return exec.CommandContext(ctx, "sleep", "5").Run()
		}

		// When the retries run out after the deadline passes
		err := retryKeepingLastError(ctx, NewTestingT(), "k3d image import", 3, 10*time.Millisecond, action)

		// Then the reported cause is the real one, not the kill that buried it
		Expect(err).To(MatchError(ContainSubstring("content digest sha256:abc: not found")))
		Expect(attempts).To(BeNumerically(">", 1))
	})

	It("does not panic when the budget is already spent", func() {
		// Given a context that is done before the first attempt
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// When it is asked to retry, which terratest answers with an unchecked cast
		err := retryKeepingLastError(ctx, NewTestingT(), "k3d image import", 2, time.Millisecond, func() error {
			return nil
		})

		// Then it reports rather than panics
		Expect(err).To(MatchError(ContainSubstring("context canceled")))
	})

	It("returns nil once an attempt succeeds", func() {
		// Given an action that fails once and then succeeds
		attempts := 0
		action := func() error {
			attempts++
			if attempts == 1 {
				return errors.New("transient")
			}
			return nil
		}

		// When it is retried
		err := retryKeepingLastError(context.Background(), NewTestingT(), "load images", 3, time.Millisecond, action)

		// Then the earlier failure is not reported
		Expect(err).ToNot(HaveOccurred())
		Expect(attempts).To(Equal(2))
	})
})
