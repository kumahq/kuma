package mtls_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EMTLS(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "MTLS tests")
	} else {
		t.SkipNow()
	}
}
