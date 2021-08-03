// +build gateway

package cmd

import (
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/kumactl"
)

func init() {
	kumactl_cmd.Plugins = append(kumactl_cmd.Plugins, &gateway.Plugin{})
}
