package v3_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_test "github.com/kumahq/kuma/pkg/util/test"
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

func TestModifications(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Modifications V3 Suite",
		[]Reporter{util_test.NewlineReporter{}})
}
