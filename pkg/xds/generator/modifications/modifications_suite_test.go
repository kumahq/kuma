package modifications_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_test "github.com/Kong/kuma/pkg/util/test"
)

func TestModifications(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Modifications Suite",
		[]Reporter{util_test.NewlineReporter{}})
}
