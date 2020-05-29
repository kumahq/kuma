package framework

import "time"

const (
	Verbose = true
	Silent  = false

	defaultKumactlConfig         = "${HOME}/.kumactl/config"
	defaultKubeConfigPathPattern = "${HOME}/.kube/kind-kuma-%d-config"
	envKUMACTLBIN                = "KUMACTLBIN"
	maxClusters                  = 3
	defaultRetries               = 10
	defaultTiemout               = 3 * time.Second

	kumaNamespace   = "kuma-system"
	kumaServiceName = "kuma-control-plane"

	kumaCPImage   = "kuma/kuma-cp"
	kumaDPImage   = "kuma/kuma-dp"
	kumaInitImage = "kuma/kuma-init"
)
