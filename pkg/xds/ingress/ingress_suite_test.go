package ingress_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_test "github.com/Kong/kuma/pkg/util/test"
)

func TestIngress(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Ingress Suite",
		[]Reporter{util_test.NewlineReporter{}})
}
