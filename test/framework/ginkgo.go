package framework

import (
	"github.com/onsi/ginkgo/v2"
)

func ShouldSkipCleanup() bool {
	suiteConfig, _ := ginkgo.GinkgoConfiguration()

	return ginkgo.CurrentSpecReport().Failed() && suiteConfig.FailFast
}

func E2EAfterEach(fn func()) {
	ginkgo.AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		fn()
	})
}

func E2EAfterSuite(fn func()) {
	ginkgo.AfterSuite(func() {
		if ShouldSkipCleanup() {
			return
		}
		fn()
	})
}

func E2EBeforeSuite(fn func()) {
	ginkgo.BeforeSuite(func() {
		fn()
	})
}
