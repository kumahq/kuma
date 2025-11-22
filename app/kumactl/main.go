package main

import (
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kumahq/kuma/v2/app/kumactl/cmd"
)

func init() {
	// Initialize controller-runtime logger before any packages are loaded
	// to prevent panic in controller-runtime/pkg/cache/internal.init()
	// This sets a default logger that will be overridden later if needed
	kube_ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
}

func main() {
	cmd.Execute()
}
