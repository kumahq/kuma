package framework

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
	kuma_version "github.com/kumahq/kuma/pkg/version"
	"github.com/kumahq/kuma/test/framework/versions"
)

var _ config.Config = E2eConfig{}

type E2eConfig struct {
	config.BaseConfig

	KumaImageRegistry                 string            `json:"imageRegistry,omitempty" envconfig:"KUMA_GLOBAL_IMAGE_REGISTRY"`
	KumaImageTag                      string            `json:"imageTag,omitempty" envconfig:"KUMA_GLOBAL_IMAGE_TAG"`
	KumaNamespace                     string            `json:"namespace,omitempty"`
	KumaServiceName                   string            `json:"serviceName,omitempty"`
	HelmChartPath                     string            `json:"helmChartPath,omitempty"`
	HelmSubChartPrefix                string            `json:"helmSubChartPrefix,omitempty"`
	HelmChartName                     string            `json:"helmChartName,omitempty"`
	HelmRepoUrl                       string            `json:"helmRepoUrl,omitempty"`
	HelmGlobalExtraYaml               string            `json:"HelmGlobalExtraYaml,omitempty"`
	CNIApp                            string            `json:"CNIApp,omitempty"`
	CNINamespace                      string            `json:"CNINamespace,omitempty"`
	CNIConf                           CniConf           `json:"CNIConf,omitempty"`
	KumaGlobalZoneSyncServiceName     string            `json:"globalZoneSyncServiceName,omitempty"`
	KumaUniversalEnvVars              map[string]string `json:"universalEnvVars,omitempty"`
	KumaZoneUniversalEnvVars          map[string]string `json:"universalZoneEnvVars,omitempty"`
	KumaK8sCtlFlags                   map[string]string `json:"k8sCtlFlags,omitempty"`
	KumaZoneK8sCtlFlags               map[string]string `json:"k8sZoneCtlFlags,omitempty"`
	DefaultObservabilityNamespace     string            `json:"observabilityNamespace,omitempty"`
	DefaultGatewayNamespace           string            `json:"gatewayNamespace,omitempty"`
	KumactlImageRepo                  string            `json:"ctlImageRepo,omitempty" envconfig:"KUMACTL_IMAGE_REPOSITORY"`
	KumaCPImageRepo                   string            `json:"cpImageRepo,omitempty" envconfig:"KUMA_CP_IMAGE_REPOSITORY"`
	KumaDPImageRepo                   string            `json:"dpImageRepo,omitempty" envconfig:"KUMA_DP_IMAGE_REPOSITORY"`
	KumaInitImageRepo                 string            `json:"initImageRepo,omitempty" envconfig:"KUMA_INIT_IMAGE_REPOSITORY"`
	KumaCNIImageRepo                  string            `json:"cniImageRepo,omitempty" envconfig:"KUMA_CNI_IMAGE_REPOSITORY"`
	KumaUniversalImageRepo            string            `json:"universalImageRepo,omitempty"`
	XDSApiVersion                     string            `json:"xdsVersion,omitempty" envconfig:"API_VERSION"`
	K8sType                           K8sType           `json:"k8sType,omitempty" envconfig:"KUMA_K8S_TYPE"`
	IPV6                              bool              `json:"ipv6,omitempty" envconfig:"IPV6"`
	UseHostnameInsteadOfIP            bool              `json:"useHostnameInsteadOfIP,omitempty" envconfig:"KUMA_USE_HOSTNAME_INSTEAD_OF_ID"`
	UseLoadBalancer                   bool              `json:"useLoadBalancer,omitempty" envconfig:"KUMA_USE_LOAD_BALANCER"`
	DefaultClusterStartupRetries      int               `json:"defaultClusterStartupRetries,omitempty" envconfig:"KUMA_DEFAULT_RETRIES"`
	DefaultClusterStartupTimeout      time.Duration     `json:"defaultClusterStartupTimeout,omitempty" envconfig:"KUMA_DEFAULT_TIMEOUT"`
	KumactlBin                        string            `json:"kumactlBin,omitempty" envconfig:"KUMACTLBIN"`
	ZoneEgressApp                     string            `json:"zoneEgressApp,omitempty" envconfig:"KUMA_ZONE_EGRESS_APP"`
	ZoneIngressApp                    string            `json:"zoneIngressApp,omitempty" envconfig:"KUMA_ZONE_INGRESS_APP"`
	Arch                              string            `json:"arch,omitempty" envconfig:"ARCH"`
	OS                                string            `json:"os,omitempty" envconfig:"OS"`
	KumaCpConfig                      KumaCpConfig      `json:"kumaCpConfig,omitempty" envconfig:"KUMA_CP_CONFIG"`
	UniversalE2ELogsPath              string            `json:"universalE2ELogsPath,omitempty" envconfig:"UNIVERSAL_E2E_LOGS_PATH"`
	CleanupLogsOnSuccess              bool              `json:"cleanupLogsOnSuccess,omitempty" envconfig:"CLEANUP_LOGS_ON_SUCCESS"`
	VersionsYamlPath                  string            `json:"versionsYamlPath,omitempty" envconfig:"VERSIONS_YAML_PATH"`
	KumaExperimentalSidecarContainers bool              `json:"kumaSidecarContainers,omitempty" envconfig:"KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS"`
	DebugDir                          string            `json:"debugDir" envconfig:"KUMA_DEBUG_DIR"`

	SuiteConfig SuiteConfig `json:"suites,omitempty"`
}

func (c E2eConfig) SupportedVersions() []versions.Version {
	return versions.ParseFromFile(c.VersionsYamlPath)
}

type SuiteConfig struct {
	Compatibility CompatibilitySuiteConfig `json:"compatibility,omitempty"`
}

type CompatibilitySuiteConfig struct {
	HelmVersion string `json:"helmVersion,omitempty"`
}

type KumaCpConfig struct {
	Standalone StandaloneConfig `json:"standalone,omitempty"`
	Multizone  MultizoneConfig  `json:"multizone,omitempty"`
}

type StandaloneConfig struct {
	Kubernetes ControlPlaneConfig `json:"kubernetes,omitempty"`
	Universal  ControlPlaneConfig `json:"universal,omitempty"`
}

type MultizoneConfig struct {
	Global    ControlPlaneConfig `json:"global,omitempty"`
	KubeZone1 ControlPlaneConfig `json:"kubeZone1,omitempty"`
	KubeZone2 ControlPlaneConfig `json:"kubeZone2,omitempty"`
	UniZone1  ControlPlaneConfig `json:"uniZone1,omitempty"`
	UniZone2  ControlPlaneConfig `json:"uniZone2,omitempty"`
}

type ControlPlaneConfig struct {
	Envs                 map[string]string `json:"envs,omitempty"`
	AdditionalYamlConfig string            `json:"additionalYamlConfig,omitempty"`
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

	Config.Arch = runtime.GOARCH
	Config.OS = runtime.GOOS

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
	VersionsYamlPath:              "../../../versions.yml",
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
		Compatibility: CompatibilitySuiteConfig{
			HelmVersion: "2.6.10",
		},
	},
	K8sType:                      KindK8sType,
	DefaultClusterStartupRetries: 30,
	DefaultClusterStartupTimeout: time.Second * 3,
	KumaCpConfig: KumaCpConfig{
		Standalone: StandaloneConfig{
			Kubernetes: ControlPlaneConfig{
				Envs: map[string]string{
					"KUMA_RUNTIME_KUBERNETES_INJECTOR_IGNORED_SERVICE_SELECTOR_LABELS": "changesvc-test-label",
				},
				AdditionalYamlConfig: "",
			},
			Universal: ControlPlaneConfig{
				Envs:                 map[string]string{},
				AdditionalYamlConfig: "",
			},
		},
		Multizone: MultizoneConfig{
			Global: ControlPlaneConfig{
				Envs:                 map[string]string{},
				AdditionalYamlConfig: "",
			},
			KubeZone1: ControlPlaneConfig{
				Envs:                 map[string]string{},
				AdditionalYamlConfig: "",
			},
			KubeZone2: ControlPlaneConfig{
				Envs:                 map[string]string{},
				AdditionalYamlConfig: "",
			},
			UniZone1: ControlPlaneConfig{
				Envs:                 map[string]string{},
				AdditionalYamlConfig: "",
			},
			UniZone2: ControlPlaneConfig{
				Envs:                 map[string]string{},
				AdditionalYamlConfig: "",
			},
		},
	},
	ZoneEgressApp:                     "kuma-egress",
	ZoneIngressApp:                    "kuma-ingress",
	UniversalE2ELogsPath:              path.Join(os.TempDir(), "e2e"),
	CleanupLogsOnSuccess:              false,
	KumaExperimentalSidecarContainers: false,
	DebugDir:                          path.Join(os.TempDir(), "e2e-debug"),
}

func init() {
	Config = defaultConf
	if err := config.Load(os.Getenv("E2E_CONFIG_FILE"), &Config); err != nil {
		panic(err)
	}

	if err := Config.AutoConfigure(); err != nil {
		panic(err)
	}
}

func KumaDeploymentOptionsFromConfig(config ControlPlaneConfig) []KumaDeploymentOption {
	kumaOptions := []KumaDeploymentOption{}
	for key, value := range config.Envs {
		kumaOptions = append(kumaOptions, WithEnv(key, value))
	}
	if config.AdditionalYamlConfig != "" {
		kumaOptions = append(kumaOptions, WithYamlConfig(config.AdditionalYamlConfig))
	}
	return kumaOptions
}
