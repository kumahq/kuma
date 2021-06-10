package helm_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/helm"
)

var _ = Describe("Test Zone and Global with Helm chart", helm.ZoneAndGlobalWithHelmChart)
