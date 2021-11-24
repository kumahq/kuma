package jobs_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/jobs"
)

var _ = Describe("Jobs on Kubernetes", jobs.Jobs)
