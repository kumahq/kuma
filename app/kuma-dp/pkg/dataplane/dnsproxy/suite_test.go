package dnsproxy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
)

func TestMetrics(t *testing.T) {
	test.RunSpecs(t, "DNS Proxy")
}

var _ = Describe("components", func() {

})
