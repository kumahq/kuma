package dnsproxy_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestMetrics(t *testing.T) {
	test.RunSpecs(t, "DNS Proxy")
}
