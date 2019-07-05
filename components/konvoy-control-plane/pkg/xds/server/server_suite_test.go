package server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_test "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/test"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Server Suite",
		[]Reporter{util_test.NewlineReporter{}})
}
