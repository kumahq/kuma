package framework

import (
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
