package install

import (
	"context"
	"encoding/base64"
	"fmt"
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

func (c *InstallerConfig) PrepareKubeconfig(token []byte, caCrtPath string) error {
	if c.KubernetesServiceHost == "" {
		return errors.New("kubernetes service host is not set")
	}

	if c.KubernetesServicePort == "" {
		return errors.New("kubernetes service port is not set")
	}

	if c.KubernetesCaFile == "" {
		c.KubernetesCaFile = caCrtPath
	}

	kubeCa, err := os.ReadFile(c.KubernetesCaFile)
	if err != nil {
		return errors.Wrap(err, "failed to read Kubernetes CA file")
	}

	return c.writeKubeconfig(c.GenerateKubeconfigTemplate(token, kubeCa))
}

func (c *InstallerConfig) writeKubeconfig(kubeconfig string) error {
	path := filepath.Join(c.MountedCniNetDir, c.KubeconfigName)

	log.Info("writing kubernetes config", "path", path)

	if err := atomic.WriteFile(path, strings.NewReader(kubeconfig)); err != nil {
		return errors.Wrap(err, "failed to write kubeconfig")
	}

	return nil
}

func (c *InstallerConfig) GenerateKubeconfigTemplate(token, caData []byte) string {
	return fmt.Sprintf(
		`# Kubeconfig file for kuma CNI plugin.
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: %s
    certificate-authority-data: %s
users:
- name: kuma-cni
  user:
    token: %s
contexts:
- name: kuma-cni-context
  context:
    cluster: local
    user: kuma-cni
current-context: kuma-cni-context`,
		c.kubernetesServiceURL(),
		base64.StdEncoding.EncodeToString(caData),
		token,
	)
}

func (c *InstallerConfig) kubernetesServiceURL() *url.URL {
	return &url.URL{
		Scheme: c.KubernetesServiceProtocol,
		Host:   net.JoinHostPort(c.KubernetesServiceHost, c.KubernetesServicePort),
	}
}

func (c *InstallerConfig) PrepareKumaCniConfig(ctx context.Context, token []byte) error {
	// Replace placeholders in the CNI network configuration
	cniConfig := strings.NewReplacer(
		"__KUBECONFIG_FILEPATH__", filepath.Join(c.HostCniNetDir, c.KubeconfigName),
		"__SERVICEACCOUNT_TOKEN__", string(token),
	).Replace(c.CniNetworkConfig)

	log.V(1).Info("CNI config after replacement", "CNI config", cniConfig)

	if c.ChainedCniPlugin {
		if err := setupChainedPlugin(ctx, c.MountedCniNetDir, c.CniConfName, cniConfig); err != nil {
			return errors.Wrap(err, "unable to setup kuma CNI as chained plugin")
		}

		return nil
	}

	configPath := filepath.Join(c.MountedCniNetDir, c.CniConfName)
	log.Info("writing standalone CNI config", "path", configPath)

	if err := atomic.WriteFile(configPath, strings.NewReader(cniConfig)); err != nil {
		return errors.Wrap(err, "failed to write standalone CNI config")
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
