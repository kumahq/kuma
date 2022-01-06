package permissive_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/mtls/permissive"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EMTLSPermissive(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "mTLS Permissive Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Permissive mTLS", permissive.PermissiveMode)
