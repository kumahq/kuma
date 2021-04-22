package helm_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/helm"
)

var _ = Describe("Test upgrading with Helm chart", helm.UpgradingWithHelmChart)
