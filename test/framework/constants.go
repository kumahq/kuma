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
)

const (
	defaultKumactlConfig             = "${HOME}/.kumactl/%s-config"
	defaultToolKubeConfigPathPattern = "${HOME}/.kube/%s-%s.yaml"
	legacyKubeConfigPathPattern      = "${HOME}/.kube/%s.yaml"
	oldKindKubeConfigPathPattern     = "${HOME}/.kube/kind-%s-config"
)

const (
	dataplaneConfigurationRefreshIntervalEnv = "KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL"
	// Below the 10s default so e2e assertions don't idle on xDS refreshes; per-test envs override it.
	e2eDataplaneConfigurationRefreshInterval = "3s"
)
