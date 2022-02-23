package universal_multizone_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficroute/universal_multizone"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Traffic Route Universal Multizone Suite")
}

var _ = Describe("Test Multizone Universal deployment", universal_multizone.KumaMultizone)
