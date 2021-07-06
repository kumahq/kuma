package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"

	"github.com/kumahq/kuma/app/kuma-cni/cmd"
)

func main() {
	skel.PluginMain(cmd.Add, cmd.Get, cmd.Del, version.All, "kuma-cni")
}
