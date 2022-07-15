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
	"k8s.io/utils/env"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/util/files"
)

const (
	defaultKumaCniConfName = "YYY-kuma-cni.conflist"
)

type InstallerConfig struct {
	HostCniNetDir    string `envconfig:"cni_net_dir" default:"/etc/cni/net.d"`
	KubeconfigName   string `envconfig:"kubecfg_file_name" default:"ZZZ-kuma-cni-kubeconfig"`
	CfgCheckInterval int    `envconfig:"cfgcheck_interval" default:"1"`
	ChainedCniPlugin bool   `envconfig:"chained_cni_plugin" default:"true"`
	MountedCniNetDir string `envconfig:"mounted_cni_net_dir" default:"/host/etc/cni/net.d"`
	ShouldSleep      bool   `envconfig:"sleep" default:"true"`
	CniConfName      string `envconfig:"cni_conf_name" default:""`
}

func (i InstallerConfig) Sanitize() {
}

func (i InstallerConfig) Validate() error {
	if i.CfgCheckInterval <= 0 {
		return errors.New("CFGCHECK_INTERVAL env variable needs to be greater than 0")
	}

	// should I check that dirs exist?

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

func prepareKubeconfig(mountedCniNetDir, kubeconfigName, serviceAccountPath string) error {
	kubeconfigPath := mountedCniNetDir + "/" + kubeconfigName
	serviceAccountTokenPath := serviceAccountPath + "/token"
	serviceAccountToken, err := ioutil.ReadFile(serviceAccountTokenPath)
	if err != nil {
		return err
	}

	if files.FileExists(serviceAccountTokenPath) {
		kubernetesServiceHost := env.GetString("KUBERNETES_SERVICE_HOST", "")
		if kubernetesServiceHost == "" {
			return errors.New("KUBERNETES_SERVICE_HOST env variable not set")
		}

		kubernetesServicePort := env.GetString("KUBERNETES_SERVICE_PORT", "")
		if kubernetesServicePort == "" {
			return errors.New("KUBERNETES_SERVICE_PORT env variable not set")
		}

		kubeCaFile := env.GetString("KUBE_CA_FILE", serviceAccountPath+"/ca.crt")
		kubeCa, err := ioutil.ReadFile(kubeCaFile)
		if err != nil {
			return err
		}
		kubernetesServiceProtocol := env.GetString("KUBERNETES_SERVICE_PROTOCOL", "https")
		caData := base64.StdEncoding.EncodeToString(kubeCa)

		kubeconfig := kubeconfigTemplate(kubernetesServiceProtocol, kubernetesServiceHost, kubernetesServicePort, string(serviceAccountToken), caData)
		log.Info("writing kubeconfig", "path", kubeconfigPath)
		err = atomic.WriteFile(kubeconfigPath, strings.NewReader(kubeconfig))
		if err != nil {
			return err
		}
	}

	return nil
}

func kubeconfigTemplate(protocol, host, port, token, caData string) string {
	safeHost := host
	if govalidator.IsIPv6(host) {
		if !(strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]")) {
			safeHost = "[" + host + "]"
		}
	}

	serverUrl := url.URL{
		Scheme: protocol,
		Host:   safeHost + ":" + port,
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
	rawConfig := env.GetString("CNI_NETWORK_CONFIG", "")
	kubeconfigFilePath := ic.HostCniNetDir + "/" + ic.KubeconfigName

	config := strings.Replace(rawConfig, "__KUBECONFIG_FILEPATH__", kubeconfigFilePath, 1)
	log.V(1).Info("config after replace", "config", config)

	serviceAccountToken, err := ioutil.ReadFile(serviceAccountPath + "/token")
	if err != nil {
		return err
	}
	config = strings.Replace(config, "__SERVICEACCOUNT_TOKEN__", string(serviceAccountToken), 1)

	if ic.ChainedCniPlugin {
		err := setupChainedPlugin(ic.MountedCniNetDir, ic.CniConfName, config)
		if err != nil {
			return errors.Wrap(err, "unable to setup kuma cni as chained plugin")
		}
	} else {
		err := atomic.WriteFile(ic.MountedCniNetDir+"/"+ic.CniConfName, strings.NewReader(config))
		if err != nil {
			return err
		}
	}

	return nil
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
			log.Info("could not find cni conf file using default")
			installerConfig.CniConfName = defaultKumaCniConfName
		} else {
			log.Info("found CNI config file", "file", cniConfFile)
			installerConfig.CniConfName = cniConfFile
		}
	}

	return &installerConfig, nil
}
