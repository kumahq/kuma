package framework

import (
	"time"
)

const (
	Verbose = true
	Silent  = false

	TestNamespace = "kuma-test"

	Kuma1 = "kuma-1"
	Kuma2 = "kuma-2"
	Kuma3 = "kuma-3"
	Kuma4 = "kuma-4"
	Kuma5 = "kuma-5"
	Kuma6 = "kuma-6"

	DefaultRetries = 30
	DefaultTimeout = 3 * time.Second

	universalKDSPort = 30686
)

const (
	defaultKumactlConfig         = "${HOME}/.kumactl/%s-config"
	defaultKubeConfigPathPattern = "${HOME}/.kube/kind-%s-config"

	redirectPortInbound   = "15006"
	redirectPortInboundV6 = "15010"
	redirectPortOutbound  = "15001"
)
