package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"

	"github.com/kumahq/kuma/app/kuma-cni/cmd"
	"github.com/kumahq/kuma/app/kuma-cni/pkg/log"
)

func main() {
	logger := log.NewLogger()
	log.Log = logger.Log

	skel.PluginMain(cmd.Add, cmd.Get, cmd.Del, version.All, "kuma-cni")
	logger.Exit(0)
}
