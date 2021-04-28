package framework

import (
	"github.com/onsi/ginkgo"
	ginkgo_config "github.com/onsi/ginkgo/config"
)

func ShouldSkipCleanup() bool {
	return ginkgo.CurrentGinkgoTestDescription().Failed && ginkgo_config.GinkgoConfig.FailFast
}
