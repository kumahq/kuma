// +build gateway

package gateway

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
)

func init() {
	core_plugins.Register("gateway", &plugin{})
}
