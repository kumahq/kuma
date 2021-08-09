// +build gateway

package gateway

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
)

func init() {
	plugins.Register("gateway", &plugin{})
}
