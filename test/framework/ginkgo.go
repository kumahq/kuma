package framework

import (
	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	ginkgo_config "github.com/onsi/ginkgo/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kumahq/kuma/pkg/core"
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
		core.SetLogger = func(l logr.Logger) {}
		logf.SetLogger(zap.LoggerTo(ginkgo.GinkgoWriter, true))
		fn()
	})
}
