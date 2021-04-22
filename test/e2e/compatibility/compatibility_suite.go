package compatibility

import (
	ginkgo_config "github.com/onsi/ginkgo/config"

	. "github.com/onsi/ginkgo"
)

func ShouldSkipCleanup() bool {
	return CurrentGinkgoTestDescription().Failed && ginkgo_config.GinkgoConfig.FailFast
}
