package tls_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func Test(t *testing.T) {
	test.RunSpecs(t, "Envoy TLS Suite")
}
