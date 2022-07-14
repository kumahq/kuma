package main

import (
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/natefinch/atomic"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/util/files"
)

const (
	defaultKumaCniConfName = "YYY-kuma-cni.conflist"
)

type InstallerConfig struct {
	CfgCheckInterval          int    `envconfig:"cfgcheck_interval" default:"1"`
	ChainedCniPlugin          bool   `envconfig:"chained_cni_plugin" default:"true"`
	CniConfName               string `envconfig:"cni_conf_name" default:""`
	CniNetworkConfig          string `envconfig:"cni_network_config" default:""`
	HostCniNetDir             string `envconfig:"cni_net_dir" default:"/etc/cni/net.d"`
	KubeconfigName            string `envconfig:"kubecfg_file_name" default:"ZZZ-kuma-cni-kubeconfig"`
	KubernetesCaFile          string `envconfig:"kube_ca_file"`
	KubernetesServiceHost     string `envconfig:"kubernetes_service_host"`
	KubernetesServicePort     string `envconfig:"kubernetes_service_port"`
	KubernetesServiceProtocol string `envconfig:"kubernetes_service_protocol" default:"https"`
	MountedCniNetDir          string `envconfig:"mounted_cni_net_dir" default:"/host/etc/cni/net.d"`
	ShouldSleep               bool   `envconfig:"sleep" default:"true"`
}

func (i InstallerConfig) Sanitize() {
}

func (i InstallerConfig) Validate() error {
	if i.CfgCheckInterval <= 0 {
		return errors.New("CFGCHECK_INTERVAL env variable needs to be greater than 0")
	}

	if ok, _ := govalidator.IsFilePath(i.HostCniNetDir); ok == false {
		return errors.New("CNI_NET_DIR must be a valid path")
	}

	if ok, _ := govalidator.IsFilePath(i.KubernetesCaFile); ok == false {
		return errors.New("KUBE_CA_FILE must be a valid path")
	}

	if !govalidator.IsHost(i.KubernetesServiceHost) {
		return errors.New("KUBERNETES_SERVICE_HOST must be a valid host")
	}

	if !govalidator.IsPort(i.KubernetesServicePort) {
		return errors.New("KUBERNETES_SERVICE_PORT must be a valid port (between 1 and 65535)")
	}

	if ok, _ := govalidator.IsFilePath(i.MountedCniNetDir); ok == false {
		return errors.New("MOUNTED_CNI_NET_DIR must be a valid path")
	}

	return nil
}

func findCniConfFile(mountedCNINetDir string) (string, error) {
	matches, err := filepath.Glob(mountedCNINetDir + "/*.conf")
	if err != nil {
		return "", err
	}

	file, found := lookForValidConfig(matches, isValidConfFile)
	if found {
		return file, nil
	}

	matches, err = filepath.Glob(mountedCNINetDir + "/*.conflist")
	if err != nil {
		return "", err
	}
	file, found = lookForValidConfig(matches, isValidConflistFile)
	if found {
		return file, nil
	}

	// use default
	return "", errors.New("cni conf file not found - use default")
}

func prepareKubeconfig(ic *InstallerConfig, serviceAccountPath string) error {
	kubeconfigPath := ic.MountedCniNetDir + "/" + ic.KubeconfigName
	serviceAccountTokenPath := serviceAccountPath + "/token"
	serviceAccountToken, err := ioutil.ReadFile(serviceAccountTokenPath)
	if err != nil {
		return err
	}

	if files.FileExists(serviceAccountTokenPath) {
		kubeCa, err := ioutil.ReadFile(ic.KubernetesCaFile)
		if err != nil {
			return err
		}
		caData := base64.StdEncoding.EncodeToString(kubeCa)

		kubeconfig := kubeconfigTemplate(ic, string(serviceAccountToken), caData)
		log.Info("writing kubeconfig", "path", kubeconfigPath)
		err = atomic.WriteFile(kubeconfigPath, strings.NewReader(kubeconfig))
		if err != nil {
			return err
		}
	}

	return nil
}

func kubeconfigTemplate(ic *InstallerConfig, token, caData string) string {
	safeHost := ic.KubernetesServiceHost
	if govalidator.IsIPv6(ic.KubernetesServiceHost) {
		if !(surroundedByBrackets(ic.KubernetesServiceHost)) {
			safeHost = "[" + ic.KubernetesServiceHost + "]"
		}
	}

	serverUrl := url.URL{
		Scheme: ic.KubernetesServiceProtocol,
		Host:   safeHost + ":" + ic.KubernetesServicePort,
	}

	return `# Kubeconfig file for kuma CNI plugin.
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: ` + serverUrl.String() + `
    certificate-authority-data: ` + caData + `
users:
- name: kuma-cni
  user:
    token: ` + token + `
contexts:
- name: kuma-cni-context
  context:
    cluster: local
    user: kuma-cni
current-context: kuma-cni-context`
}

func prepareKumaCniConfig(ic *InstallerConfig, serviceAccountPath string) error {
	rawConfig := ic.CniNetworkConfig
	kubeconfigFilePath := ic.HostCniNetDir + "/" + ic.KubeconfigName

	kumaConfig := strings.Replace(rawConfig, "__KUBECONFIG_FILEPATH__", kubeconfigFilePath, 1)
	log.V(1).Info("kumaConfig after replace", "kumaConfig", kumaConfig)

	serviceAccountToken, err := ioutil.ReadFile(serviceAccountPath + "/token")
	if err != nil {
		return err
	}
	kumaConfig = strings.Replace(kumaConfig, "__SERVICEACCOUNT_TOKEN__", string(serviceAccountToken), 1)

	if ic.ChainedCniPlugin {
		err := setupChainedPlugin(ic.MountedCniNetDir, ic.CniConfName, kumaConfig)
		if err != nil {
			return errors.Wrap(err, "unable to setup kuma cni as chained plugin")
		}
	}

	return nil
}

func surroundedByBrackets(text string) bool {
	return strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]")
}

func loadInstallerConfig() (*InstallerConfig, error) {
	var installerConfig InstallerConfig
	err := config.Load("", &installerConfig)
	if err != nil {
		return nil, err
	}

	if installerConfig.CniConfName == "" {
		cniConfFile, err := findCniConfFile(installerConfig.MountedCniNetDir)
		if err != nil {
			log.Error(err, "could not find cni conf file using default")
			installerConfig.CniConfName = defaultKumaCniConfName
		}
		installerConfig.CniConfName = cniConfFile
	}

	return &installerConfig, nil
}
