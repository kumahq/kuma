package projectedsatoken_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/projectedsatoken"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Projected SAT")
}

var _ = Describe("Test Projected Service Account Token on Universal", Label("job-0"), projectedsatoken.ProjectedServiceAccountToken)
