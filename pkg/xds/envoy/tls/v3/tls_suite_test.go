package v3_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_test "github.com/kumahq/kuma/pkg/util/test"
)

func TestTLS(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Envoy TLS Suite",
		[]Reporter{util_test.NewlineReporter{}})
}
