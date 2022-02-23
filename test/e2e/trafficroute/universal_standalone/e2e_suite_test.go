package universal_standalone_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficroute/universal_standalone"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Traffic Route Universal Standalone Suite")
}

var _ = Describe("Test Standalone Universal deployment", universal_standalone.KumaStandalone)
