package transparentproxy_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/transparentproxy"
)

func TestTransparentProxy(t *testing.T) {
	transparentproxy.InitConfig()
	test.RunSpecs(t, "Transparent Proxy Suite")
}
