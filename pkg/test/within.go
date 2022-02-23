package test

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Within returns a function that executes the given test task in
// a dedicated goroutine, asserting that it must complete within
// the given timeout.
//
// See https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#removed-async-testing
func Within(timeout time.Duration, task func()) func() {
	return func() {
		done := make(chan interface{})

		go func() {
			defer ginkgo.GinkgoRecover()
			defer close(done)
			task()
		}()

		gomega.Eventually(done, timeout).Should(gomega.BeClosed())
	}
}
