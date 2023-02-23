package framework

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pkg/errors"

	core_config "github.com/kumahq/kuma/pkg/config"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type E2eConfig struct {
	KumaImageRegistry             string            `json:"imageRegistry,omitempty" envconfig:"KUMA_GLOBAL_IMAGE_REGISTRY"`
	KumaImageTag                  string            `json:"imageTag,omitempty" envconfig:"KUMA_GLOBAL_IMAGE_TAG"`
	KumaNamespace                 string            `json:"namespace,omitempty"`
	KumaServiceName               string            `json:"serviceName,omitempty"`
	HelmChartPath                 string            `json:"helmChartPath,omitempty"`
	HelmSubChartPrefix            string            `json:"helmSubChartPrefix,omitempty"`
	HelmChartName                 string            `json:"helmChartName,omitempty"`
	HelmRepoUrl                   string            `json:"helmRepoUrl,omitempty"`
	HelmGlobalExtraYaml           string            `json:"HelmGlobalExtraYaml,omitempty"`
	CNIApp                        string            `json:"CNIApp,omitempty"`
	CNINamespace                  string            `json:"CNINamespace,omitempty"`
	CNIConf                       CniConf           `json:"CNIConf,omitempty"`
	KumaGlobalZoneSyncServiceName string            `json:"globalZoneSyncServiceName,omitempty"`
	KumaUniversalEnvVars          map[string]string `json:"universalEnvVars,omitempty"`
	KumaZoneUniversalEnvVars      map[string]string `json:"universalZoneEnvVars,omitempty"`
	KumaK8sCtlFlags               map[string]string `json:"k8sCtlFlags,omitempty"`
	KumaZoneK8sCtlFlags           map[string]string `json:"k8sZoneCtlFlags,omitempty"`
	DefaultObservabilityNamespace string            `json:"observabilityNamespace,omitempty"`
	DefaultGatewayNamespace       string            `json:"gatewayNamespace,omitempty"`
	KumactlImageRepo              string            `json:"ctlImageRepo,omitempty" envconfig:"KUMACTL_IMAGE_REPOSITORY"`
	KumaCPImageRepo               string            `json:"cpImageRepo,omitempty" envconfig:"KUMA_CP_IMAGE_REPOSITORY"`
	KumaDPImageRepo               string            `json:"dpImageRepo,omitempty" envconfig:"KUMA_DP_IMAGE_REPOSITORY"`
	KumaInitImageRepo             string            `json:"initImageRepo,omitempty" envconfig:"KUMA_INIT_IMAGE_REPOSITORY"`
	KumaCNIImageRepo              string            `json:"cniImageRepo,omitempty" envconfig:"KUMA_CNI_IMAGE_REPOSITORY"`
	KumaUniversalImageRepo        string            `json:"universalImageRepo,omitempty"`
	XDSApiVersion                 string            `json:"xdsVersion,omitempty" envconfig:"API_VERSION"`
	K8sType                       K8sType           `json:"k8sType,omitempty" envconfig:"KUMA_K8S_TYPE"`
	IPV6                          bool              `json:"ipv6,omitempty" envconfig:"IPV6"`
	UseHostnameInsteadOfIP        bool              `json:"useHostnameInsteadOfIP,omitempty" envconfig:"KUMA_USE_HOSTNAME_INSTEAD_OF_ID"`
	UseLoadBalancer               bool              `json:"useLoadBalancer,omitempty" envconfig:"KUMA_USE_LOAD_BALANCER"`
	CIDR                          string            `json:"kumaCidr,omitempty"`
	DefaultClusterStartupRetries  int               `json:"defaultClusterStartupRetries,omitempty" envconfig:"KUMA_DEFAULT_RETRIES"`
	DefaultClusterStartupTimeout  time.Duration     `json:"defaultClusterStartupTimeout,omitempty" envconfig:"KUMA_DEFAULT_TIMEOUT"`
	KumactlBin                    string            `json:"kumactlBin,omitempty" envconfig:"KUMACTLBIN"`
	ZoneEgressApp                 string            `json:"zoneEgressApp,omitempty" envconfig:"KUMA_ZONE_EGRESS_APP"`
	ZoneIngressApp                string            `json:"zoneIngressApp,omitempty" envconfig:"KUMA_ZONE_INGRESS_APP"`
	Arch                          string            `json:"arch,omitempty" envconfig:"ARCH"`

	SuiteConfig SuiteConfig `json:"suites,omitempty"`
}

type SuiteConfig struct {
	Compatibility CompatibilitySuiteConfig `json:"compatibility,omitempty"`
	Helm          HelmSuiteConfig          `json:"helm,omitempty"`
}

type CompatibilitySuiteConfig struct {
	HelmVersion string `json:"helmVersion,omitempty"`
}

type HelmSuiteConfig struct {
	Versions []string `json:"versions,omitempty"`
}

func (c E2eConfig) Sanitize() {
}

func (c E2eConfig) Validate() error {
	if Config.KumactlBin != "" {
		_, err := os.Stat(Config.KumactlBin)
		if os.IsNotExist(err) {
			return errors.Wrapf(err, "unable to find kumactl at:%s", Config.KumactlBin)
		}
	}
	return nil
}

func (c E2eConfig) AutoConfigure() error {
	if Config.CNIConf.ConfName == "" {
		switch Config.K8sType {
		case KindK8sType:
			Config.CNIConf = CniConf{
				ConfName: "10-kindnet.conflist",
				NetDir:   "/etc/cni/net.d",
				BinDir:   "/opt/cni/bin",
			}
		case K3dK8sType:
			Config.CNIConf = CniConf{
				ConfName: "10-flannel.conflist",
				NetDir:   "/var/lib/rancher/k3s/agent/etc/cni/net.d",
				BinDir:   "/bin",
			}
		case K3dCalicoK8sType:
			Config.CNIConf = CniConf{
				ConfName: "10-calico.conflist",
				NetDir:   "/etc/cni/net.d/",
				BinDir:   "/opt/cni/bin",
			}
		case AzureK8sType:
			Config.CNIConf = CniConf{
				ConfName: "10-azure.conflist",
				NetDir:   "/etc/cni/net.d",
				BinDir:   "/opt/cni/bin",
			}
		case AwsK8sType:
			Config.CNIConf = CniConf{
				ConfName: "10-aws.conflist",
				NetDir:   "/etc/cni/net.d",
				BinDir:   "/opt/cni/bin",
			}
		default:
			return fmt.Errorf("you must set a supported KUMA_K8S_TYPE got:%s", Config.K8sType)
		}
	}
	if Config.IPV6 && Config.CIDR == "" {
		Config.CIDR = "fd00:fd00::/64"
	}
	Config.Arch = runtime.GOARCH
	return nil
}

type K8sType string

const (
	KindK8sType      K8sType = "kind"
	K3dK8sType       K8sType = "k3d"
	K3dCalicoK8sType K8sType = "k3d-calico"
	AzureK8sType     K8sType = "azure"
	AwsK8sType       K8sType = "aws"
)

type CniConf struct {
	BinDir   string
	NetDir   string
	ConfName string
}

var Config E2eConfig

func (c E2eConfig) GetUniversalImage() string {
	if c.KumaImageTag != "" {
		return fmt.Sprintf("%s/%s:%s", c.KumaImageRegistry, c.KumaUniversalImageRepo, c.KumaImageTag)
	} else {
		return fmt.Sprintf("%s/%s", c.KumaImageRegistry, c.KumaUniversalImageRepo)
	}
}

var defaultConf = E2eConfig{
	HelmChartName:                 "kuma/kuma",
	HelmChartPath:                 "../../../deployments/charts/kuma",
	HelmRepoUrl:                   "https://kumahq.github.io/charts",
	HelmSubChartPrefix:            "",
	KumaNamespace:                 "kuma-system",
	KumaServiceName:               "kuma-control-plane",
	KumaGlobalZoneSyncServiceName: "kuma-global-zone-sync",
	DefaultObservabilityNamespace: "mesh-observability",
	DefaultGatewayNamespace:       "kuma-gateway",
	CNIApp:                        "kuma-cni",
	CNINamespace:                  "kube-system",

	KumaImageRegistry:      "kumahq",
	KumaImageTag:           kuma_version.Build.Version,
	KumaUniversalImageRepo: "kuma-universal",
	KumaCPImageRepo:        "kuma-cp",
	KumaDPImageRepo:        "kuma-dp",
	KumaInitImageRepo:      "kuma-init",
	KumactlImageRepo:       "kumactl",
	KumaCNIImageRepo:       "kuma-cni",

	KumaUniversalEnvVars: map[string]string{},
	KumaK8sCtlFlags:      map[string]string{},
	KumaZoneK8sCtlFlags:  map[string]string{},
	SuiteConfig: SuiteConfig{
		Helm: HelmSuiteConfig{
			Versions: []string{
				"1.6.0",
			},
		},
		Compatibility: CompatibilitySuiteConfig{
			HelmVersion: "1.6.0",
		},
	},
	K8sType:                      KindK8sType,
	DefaultClusterStartupRetries: 30,
	DefaultClusterStartupTimeout: time.Second * 3,

	ZoneEgressApp:  "kuma-egress",
	ZoneIngressApp: "kuma-ingress",
}

func init() {
	Config = defaultConf
	if err := core_config.Load(os.Getenv("E2E_CONFIG_FILE"), &Config); err != nil {
		panic(err)
	}

	if err := Config.AutoConfigure(); err != nil {
		panic(err)
	}
}
