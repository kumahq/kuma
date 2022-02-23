package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/auth"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Auth Suite")
}

var _ = Describe("Test Universal", auth.AuthUniversal)
