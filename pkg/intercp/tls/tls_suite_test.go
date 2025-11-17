package tls_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestTLS(t *testing.T) {
	test.RunSpecs(t, "InterCP TLS Suite")
}
