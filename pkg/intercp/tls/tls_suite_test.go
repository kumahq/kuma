package tls_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestTLS(t *testing.T) {
	test.RunSpecs(t, "InterCP TLS Suite")
}
