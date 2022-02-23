package framework

import (
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
)

type TestingT struct {
	ginkgo.GinkgoTInterface
	desc ginkgo.SpecReport
}

func NewTestingT() *TestingT {
	return &TestingT{ginkgo.GinkgoT(), ginkgo.CurrentSpecReport()}
}

func (i *TestingT) Helper() {

}

func (i *TestingT) Name() string {
	return i.desc.FullText()
}

// Logf logs a test progress message.
func Logf(format string, args ...interface{}) {
	logger.Default.Logf(ginkgo.GinkgoT(), format, args...)
}
