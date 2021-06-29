package installer

const (
	MountedCNINetDir     = "mounted-cni-net-dir"
	CNINetDir            = "cni-net-dir"
	CNIConfName          = "cni-conf-name"
	ChainedCNIPlugin     = "chained-cni-plugin"
	CNINetworkConfigFile = "cni-network-config-file"
	CNINetworkConfig     = "cni-network-config"
	LogLevel             = "log-level"
	KubeconfigFilename   = "kubecfg-file-name"
	KubeconfigMode       = "kubeconfig-mode"
	KubeCAFile           = "kube-ca-file"
	SkipTLSVerify        = "skip-tls-verify"

	CNIBinDir             = "/opt/cni/bin"
	HostCNIBinDir         = "/host/opt/cni/bin"
	ServiceAccountPath    = "/var/run/secrets/kubernetes.io/serviceaccount"
	PrivateFileMode       = 0o600
	DefaultKubeconfigMode = PrivateFileMode
)
