package transparentproxy_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/transparentproxy"
)

func TestTransparentProxy(t *testing.T) {
	transparentproxy.InitConfig()
	test.RunSpecs(t, "Transparent Proxy Suite")
}
