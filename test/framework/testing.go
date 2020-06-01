package framework

import (
	. "github.com/onsi/ginkgo"
)

type TestingT struct {
	GinkgoTInterface
	desc GinkgoTestDescription
}

func NewTestingT() *TestingT {
	return &TestingT{GinkgoT(), CurrentGinkgoTestDescription()}
}

func (i *TestingT) Helper() {

}
func (i *TestingT) Name() string {
	return i.desc.FullTestText
}
