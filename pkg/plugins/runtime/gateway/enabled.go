// +build gateway

package gateway

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
)

func init() {
	core_plugins.Register("gateway", &plugin{})
}
