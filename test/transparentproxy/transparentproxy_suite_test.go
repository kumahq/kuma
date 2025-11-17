package transparentproxy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/transparentproxy"
	"github.com/kumahq/kuma/v2/test/transparentproxy/install"
)

func TestTransparentProxy(t *testing.T) {
	transparentproxy.InitConfig()
	test.RunSpecs(t, "Transparent Proxy Suite")
}

var _ = Describe("Install", install.Install)
