package transparentproxy_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/test/transparentproxy"
)

func TestTransparentProxy(t *testing.T) {
	transparentproxy.InitConfig()
	test.RunSpecs(t, "Transparent Proxy Suite")
}
