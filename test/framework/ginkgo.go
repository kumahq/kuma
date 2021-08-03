package framework

import (
	"github.com/onsi/ginkgo"
	ginkgo_config "github.com/onsi/ginkgo/config"
)

func ShouldSkipCleanup() bool {
	return ginkgo.CurrentGinkgoTestDescription().Failed && ginkgo_config.GinkgoConfig.FailFast
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
