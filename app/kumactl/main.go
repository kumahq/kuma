package main

import (
	kube_ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kumahq/kuma/v2/app/kumactl/cmd"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

func init() {
	// Initialize controller-runtime logger early to prevent timeout warning.
	// This logger uses a global atomic level that can be changed dynamically
	// via kuma_log.SetGlobalLogLevel() without replacing the logger instance.
	// This ensures controller-runtime components (cache, etc.) respect the
	// user's --log-level flag set in PersistentPreRunE.
	kube_ctrl.SetLogger(kuma_log.NewLoggerWithGlobalLevel())
}

func main() {
	cmd.Execute()
}
