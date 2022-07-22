package install

const kumaCniConfig = `{
	"type": "kuma-cni",
	"log_level": "info",
	"kubernetes": {
		"kubeconfig": "/etc/cni/net.d/ZZZ-kuma-cni-kubeconfig",
		"cni_bin_dir": "/opt/cni/bin",
		"exclude_namespaces": [
			"kuma-system"
		]
	}
}`

const kumaCniConfigTemplate = `{
	"type": "kuma-cni",
	"log_level": "info",
	"kubernetes": {
		"kubeconfig": "__KUBECONFIG_FILEPATH__",
		"cni_bin_dir": "/opt/cni/bin",
		"exclude_namespaces": [
			"kuma-system"
		]
	}
}`
