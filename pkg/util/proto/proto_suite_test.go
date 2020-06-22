package proto_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_test "github.com/Kong/kuma/pkg/util/test"
)

func TestProtoUtils(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Generator Suite",
		[]Reporter{util_test.NewlineReporter{}})
}
