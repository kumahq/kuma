package watchnamespaces_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/watchnamespaces"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Watch Namespaces Kubernetes Suite")
}

var _ = Describe("Watch Namespaces on Kubernetes", Label("job-0"), watchnamespaces.WatchOnlyDefinedNamespaces, Ordered)
