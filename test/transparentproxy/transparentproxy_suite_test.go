package transparentproxy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/transparentproxy"
	"github.com/kumahq/kuma/test/transparentproxy/install"
)

func TestTransparentProxy(t *testing.T) {
	transparentproxy.InitConfig()
	test.RunSpecs(t, "Transparent Proxy Suite")
}

var _ = Describe("Install", install.Install)
