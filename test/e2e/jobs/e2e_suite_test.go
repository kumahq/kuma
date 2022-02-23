package jobs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/jobs"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Jobs Kubernetes Suite")
}

var _ = Describe("Jobs on Kubernetes", jobs.Jobs)
