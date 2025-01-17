package federation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/federation"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "Federation Suite")
}

var (
	// TODO(bartsmykla): disabled while investingating flake (https://github.com/kumahq/kuma/issues/12142)
	_ = XDescribe("Federation with Kube Global", Label("job-3"), federation.FederateKubeZoneCPToKubeGlobal, Ordered)
	_ = XDescribe("Federation with Universal Global", Label("job-3"), federation.FederateKubeZoneCPToUniversalGlobal, Ordered)
)
