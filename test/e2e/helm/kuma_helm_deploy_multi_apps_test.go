package helm_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/helm"
)

var _ = FDescribe("Test App deployment with Helm chart", helm.AppDeploymentWithHelmChart)
