package bootstrap_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/bootstrap"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "Bootstrap Suite")
}

var _ = Describe("Corefile Template", Label("job-0"), bootstrap.CorefileTemplate, Ordered)
