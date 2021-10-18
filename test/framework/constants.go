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
	Kuma6 = "kuma-6"

	DefaultRetries = 30
	DefaultTimeout = 3 * time.Second

	KumaUniversalImage = "kuma-universal"

	kdsPort             = 30685
	loadBalancerKdsPort = 5685
)

const (
	defaultKumactlConfig         = "${HOME}/.kumactl/%s-config"
	defaultKubeConfigPathPattern = "${HOME}/.kube/kind-%s-config"

	envK8SCLUSTERS = "K8SCLUSTERS"
	envAPIVersion  = "API_VERSION"
	envIPv6        = "IPV6"

	maxClusters = 4

	confPath = "/kuma/kuma-cp.conf"

	kumaCPAPIPort        = 5681
	kumaCPAPIPortFwdBase = 32000 + kumaCPAPIPort

	redirectPortInbound   = "15006"
	redirectPortInboundV6 = "15010"
	redirectPortOutbound  = "15001"
	cidrIPv6              = "fd00:fd00::/64"
)

var HelmRepo = "kuma/kuma"
var HelmChartPath = "../../../deployments/charts/kuma"
var HelmSubChartPrefix = ""

var KumaNamespace = "kuma-system"
var KumaServiceName = "kuma-control-plane"
var KumaGlobalZoneSyncServiceName = "kuma-global-zone-sync"
var DefaultTracingNamespace = "kuma-tracing"

var CNIApp = "kuma-cni"
var CNINamespace = "kube-system"

var KumaImageRegistry = "kumahq"
var KumaCPImageRepo = "kuma-cp"
var KumaDPImageRepo = "kuma-dp"
var KumaInitImageRepo = "kuma-init"

var KumaUniversalDeployOpts []KumaDeploymentOption
var KumaK8sDeployOpts []KumaDeploymentOption
var KumaZoneK8sDeployOpts []KumaDeploymentOption
