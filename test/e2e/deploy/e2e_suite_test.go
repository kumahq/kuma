package deploy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/deploy"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Deploy Suite")
}

var _ = Describe("Test Universal deployment", Label("job-1"), deploy.UniversalDeployment)
