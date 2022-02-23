package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/mtls/universal"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E mTLS Universal Suite")
}

var _ = Describe("mTLS on Universal", universal.MTLSUniversal)
