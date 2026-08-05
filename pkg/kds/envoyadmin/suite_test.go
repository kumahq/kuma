package envoyadmin_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestEnvoyAdmin(t *testing.T) {
	test.RunSpecs(t, "KDS Envoy Admin Suite")
}
