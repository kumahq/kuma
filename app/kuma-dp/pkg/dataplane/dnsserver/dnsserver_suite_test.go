package dnsserver

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestEnvoy(t *testing.T) {
	test.RunSpecs(t, "DNS Server Suite")
}
