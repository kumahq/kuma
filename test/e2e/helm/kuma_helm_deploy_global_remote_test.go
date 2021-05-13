package helm_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/helm"
)

var _ = Describe("Test Remote and Global with Helm chart", helm.RemoteAndGlobalWithHelmChart)
