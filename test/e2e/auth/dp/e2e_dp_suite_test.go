package dp_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/auth/dp"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Auth DP Suite")
}

var _ = Describe("Test Universal", dp.DpAuthUniversal)
