package proxytemplate_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/proxytemplate"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Proxy Template Suite")
}

var _ = Describe("Proxy Template on Universal", proxytemplate.ProxyTemplateUniversal)
