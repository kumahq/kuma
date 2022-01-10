package deploy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/deploy"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EDeploy(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Deploy Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Zone and Global", deploy.ZoneAndGlobal)
var _ = Describe("Test Universal deployment", deploy.UniversalDeployment)
var _ = Describe("Test Universal Transparent Proxy deployment", deploy.UniversalTransparentProxyDeployment)
