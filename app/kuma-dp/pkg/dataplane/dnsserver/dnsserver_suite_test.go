package dnsserver

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestEnvoy(t *testing.T) {
	test.RunSpecs(t, "DNS Server Suite")
}
