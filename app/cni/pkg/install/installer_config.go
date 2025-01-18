package install

import (
	"context"
	"encoding/base64"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/atomic"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

const (
	defaultKumaCniConfName = "YYY-kuma-cni.conflist"
)

var _ config.Config = &InstallerConfig{}

type InstallerConfig struct {
	config.BaseConfig

	CfgCheckInterval          int    `envconfig:"cfgcheck_interval" default:"1"`
	ChainedCniPlugin          bool   `envconfig:"chained_cni_plugin" default:"true"`
	CniConfName               string `envconfig:"cni_conf_name" default:""`
	CniLogLevel               string `envconfig:"cni_log_level" default:"info"`
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

func (c *InstallerConfig) Validate() error {
	if c.CfgCheckInterval <= 0 {
		return errors.New("CFGCHECK_INTERVAL env variable needs to be greater than 0")
	}

	// TODO: improve validation

	return nil
}

func (c *InstallerConfig) PostProcess() error {
	if c.CniConfName != "" {
		return nil
	}

	for _, ext := range []string{"*.conf", "*.conflist"} {
		matches, err := filepath.Glob(filepath.Join(c.MountedCniNetDir, ext))
		if err != nil {
			log.Info("failed to search for CNI config files", "error", err)
			continue
		}

		if file, ok := lookForValidConfig(matches, isValidConfFile); ok {
			log.Info("found CNI config file", "file", file)
			c.CniConfName = filepath.Base(file)
			return nil
		}
	}

	log.Info("could not find CNI config file, using default")
	c.CniConfName = defaultKumaCniConfName

	return nil
}

func prepareKubeconfig(ic *InstallerConfig, token string, caCrt string) error {
	kubeconfigPath := ic.MountedCniNetDir + "/" + ic.KubeconfigName
	serviceAccountToken, err := os.ReadFile(token)
	if err != nil {
		return err
	}

	if ic.KubernetesServiceHost == "" {
		return errors.New("KUBERNETES_SERVICE_HOST env variable not set")
	}

	if ic.KubernetesServicePort == "" {
		return errors.New("KUBERNETES_SERVICE_PORT env variable not set")
	}

	if ic.KubernetesCaFile == "" {
		ic.KubernetesCaFile = caCrt
	}

	kubeCa, err := os.ReadFile(ic.KubernetesCaFile)
	if err != nil {
		return err
	}
	caData := base64.StdEncoding.EncodeToString(kubeCa)

	kubeconfig := kubeconfigTemplate(ic.KubernetesServiceProtocol, ic.KubernetesServiceHost, ic.KubernetesServicePort, string(serviceAccountToken), caData)
	log.Info("writing kubernetes config", "path", kubeconfigPath)
	err = atomic.WriteFile(kubeconfigPath, strings.NewReader(kubeconfig))
	if err != nil {
		return err
	}

	return nil
}

func kubeconfigTemplate(protocol, host, port, token, caData string) string {
	serverUrl := url.URL{
		Scheme: protocol,
		Host:   net.JoinHostPort(host, port),
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

func prepareKumaCniConfig(ctx context.Context, ic *InstallerConfig, token string) error {
	rawConfig := ic.CniNetworkConfig
	kubeconfigFilePath := ic.HostCniNetDir + "/" + ic.KubeconfigName

	cniConfig := strings.Replace(rawConfig, "__KUBECONFIG_FILEPATH__", kubeconfigFilePath, 1)
	log.V(1).Info("cni config after replace", "cni config", cniConfig)

	serviceAccountToken, err := os.ReadFile(token)
	if err != nil {
		return err
	}
	cniConfig = strings.Replace(cniConfig, "__SERVICEACCOUNT_TOKEN__", string(serviceAccountToken), 1)

	if ic.ChainedCniPlugin {
		err := setupChainedPlugin(ctx, ic.MountedCniNetDir, ic.CniConfName, cniConfig)
		if err != nil {
			return errors.Wrap(err, "unable to setup kuma cni as chained plugin")
		}
	} else {
		err := atomic.WriteFile(ic.MountedCniNetDir+"/"+ic.CniConfName, strings.NewReader(cniConfig))
		if err != nil {
			return err
		}
	}

	return nil
}

func loadInstallerConfig() (*InstallerConfig, error) {
	var installerConfig InstallerConfig

	if err := config.Load("", &installerConfig); err != nil {
		return nil, err
	}

	return &installerConfig, nil
}
