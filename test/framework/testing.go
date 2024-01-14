package framework

import (
	"hash/fnv"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/util/rand"
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

func (i *TestingT) Hash() string {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(i.Name()))
	return rand.SafeEncodeString(string(hash.Sum(nil)))
}

// Logf logs a test progress message.
func Logf(format string, args ...interface{}) {
	logger.Default.Logf(ginkgo.GinkgoT(), format, args...)
}
