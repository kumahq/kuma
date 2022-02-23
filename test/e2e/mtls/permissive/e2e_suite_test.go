package permissive_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/mtls/permissive"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E mTLS Permissive Suite")
}

var _ = Describe("Test Permissive mTLS", permissive.PermissiveMode)
