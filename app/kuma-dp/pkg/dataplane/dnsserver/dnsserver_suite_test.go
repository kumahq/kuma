package dnsserver

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestEnvoy(t *testing.T) {
	test.RunSpecs(t, "DNS Server Suite")
}
