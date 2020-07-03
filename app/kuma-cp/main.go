package main

import (
	"github.com/Kong/kuma/app/kuma-cp/cmd"

	// todo add all imports from go-control-plane, otherwise validator and parser does not know about the type. Find a common place for this and import here and in kuma-dp.
	_ "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/direct_response/v2"
	_ "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/echo/v2"
	_ "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
)

func main() {
	cmd.Execute()
}
