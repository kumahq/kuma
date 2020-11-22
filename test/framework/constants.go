package framework

import "time"

const (
	Verbose = true
	Silent  = false

	TestNamespace = "kuma-test"

	Kuma1 = "kuma-1"
	Kuma2 = "kuma-2"
	Kuma3 = "kuma-3"
	Kuma4 = "kuma-4"
	Kuma5 = "kuma-5"

	DefaultRetries = 30
	DefaultTimeout = 3 * time.Second

	KumaUniversalImage = "kuma-universal"

	kdsPort = 30685
)

const (
	defaultKumactlConfig         = "${HOME}/.kumactl/%s-config"
	defaultKubeConfigPathPattern = "${HOME}/.kube/kind-%s-config"

	envKUMACTLBIN  = "KUMACTLBIN"
	envK8SCLUSTERS = "K8SCLUSTERS"

	maxClusters = 3

	confPath = "/kuma/kuma-cp.conf"

	kumaCPAPIPort        = 5681
	kumaCPAPIPortFwdBase = 32000 + kumaCPAPIPort

	cniApp       = "kuma-cni"
	cniNamespace = "kube-system"
)

var HelmChartPath = "../../deployments/charts/kuma"
var HelmSubChartPrefix = ""

var KumaNamespace = "kuma-system"
var KumaServiceName = "kuma-control-plane"

var KumaImageRegistry = "kuma"
var KumaCPImageRepo = "kuma-cp"
var KumaDPImageRepo = "kuma-dp"
var KumaInitImageRepo = "kuma-init"
