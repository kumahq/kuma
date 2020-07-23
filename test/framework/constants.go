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

	kdsPort = 30685

	ZoneTemplateK8s = `
apiVersion: kuma.io/v1alpha1
kind: Zone
mesh: default
metadata:
  name: %s
spec:
  remoteControlPlane:
    address: %s
  ingress:
    address: %s
`
	ZoneTemplateUniversal = `
type: Zone
mesh: default
name: %s
remoteControlPlane:
  address: %s
ingress:
  address: %s
`
)

const (
	defaultKumactlConfig         = "${HOME}/.kumactl/%s-config"
	defaultKubeConfigPathPattern = "${HOME}/.kube/kind-%s-config"

	envKUMACTLBIN  = "KUMACTLBIN"
	envK8SCLUSTERS = "K8SCLUSTERS"

	maxClusters = 3

	kumaNamespace   = "kuma-system"
	kumaServiceName = "kuma-control-plane"

	kumaCPImage   = "kuma/kuma-cp"
	kumaDPImage   = "kuma/kuma-dp"
	kumaInitImage = "kuma/kuma-init"

	confPath = "/kuma/kuma-cp.conf"

	kumaCPAPIPort        = 5681
	kumaCPGUIPort        = 5683
	kumaCPAPIPortFwdBase = 32000 + kumaCPAPIPort
)
