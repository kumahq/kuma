package framework

import "time"

const (
	Verbose = true
	Silent  = false

	defaultKumactlConfig         = "${HOME}/.kumactl/%s-config"
	defaultKubeConfigPathPattern = "${HOME}/.kube/kind-%s-config"

	envKUMACTLBIN  = "KUMACTLBIN"
	envK8SCLUSTERS = "K8SCLUSTERS"

	maxClusters    = 3
	defaultRetries = 30
	defaultTimeout = 3 * time.Second

	kumaNamespace   = "kuma-system"
	kumaServiceName = "kuma-control-plane"

	kumaCPImage   = "kuma/kuma-cp"
	kumaDPImage   = "kuma/kuma-dp"
	kumaInitImage = "kuma/kuma-init"

	Kuma1 = "kuma-1"
	Kuma2 = "kuma-2"
	Kuma3 = "kuma-3"

	kumaCPAPIPort       = 5681
	kumaCPAPIPortFwdLow = 32000 + kumaCPAPIPort
	kumaCPAPIPortFwdHi  = 42000 + kumaCPAPIPort
)
