package dnsproxy_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestDNSProxy(t *testing.T) {
	test.RunSpecs(t, "DNS Proxy")
}
