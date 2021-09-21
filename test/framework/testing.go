package framework

import (
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo"
)

type TestingT struct {
	ginkgo.GinkgoTInterface
	desc ginkgo.GinkgoTestDescription
}

func NewTestingT() *TestingT {
	return &TestingT{ginkgo.GinkgoT(), ginkgo.CurrentGinkgoTestDescription()}
}

func (i *TestingT) Helper() {

}

func (i *TestingT) Name() string {
	return i.desc.FullTestText
}

// Logf logs a test progress message.
func Logf(format string, args ...interface{}) {
	logger.Default.Logf(ginkgo.GinkgoT(), format, args...)
}
