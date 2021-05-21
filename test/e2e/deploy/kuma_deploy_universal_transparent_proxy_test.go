package deploy_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/deploy"
)

var _ = Describe("Test Universal Transparent Proxy deployment", deploy.UniversalTransparentProxyDeployment)
