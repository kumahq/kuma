package context

import (
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/deployments"
	"github.com/kumahq/kuma/pkg/config/core"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type InstallControlPlaneArgs struct {
	Namespace                                    string
	ControlPlane_image_pullPolicy                string            `helm:"controlPlane.image.pullPolicy"`
	ControlPlane_image_registry                  string            `helm:"controlPlane.image.registry,omitempty"`
	ControlPlane_image_repository                string            `helm:"controlPlane.image.repository"`
	ControlPlane_image_tag                       string            `helm:"controlPlane.image.tag"`
	ControlPlane_service_name                    string            `helm:"controlPlane.service.name"`
	ControlPlane_tls_general_secret              string            `helm:"controlPlane.tls.general.secretName"`
	ControlPlane_tls_general_ca_secret           string            `helm:"controlPlane.tls.general.caSecretName"`
	ControlPlane_tls_general_caBundle            string            `helm:"controlPlane.tls.general.caBundle"`
	ControlPlane_tls_apiServer_secret            string            `helm:"controlPlane.tls.apiServer.secretName"`
	ControlPlane_tls_apiServer_clientCertsSecret string            `helm:"controlPlane.tls.apiServer.clientCertsSecretName"`
	ControlPlane_tls_kdsGlobalServer_secret      string            `helm:"controlPlane.tls.kdsGlobalServer.secretName"`
	ControlPlane_tls_kdsZoneClient_secret        string            `helm:"controlPlane.tls.kdsZoneClient.secretName"`
	ControlPlane_injectorFailurePolicy           string            `helm:"controlPlane.injectorFailurePolicy"`
	ControlPlane_secrets                         []ImageEnvSecret  `helm:"controlPlane.secrets"`
	ControlPlane_envVars                         map[string]string `helm:"controlPlane.envVars"`
	ControlPlane_nodeSelector                    map[string]string `helm:"controlPlane.nodeSelector"`
	DataPlane_image_registry                     string            `helm:"dataPlane.image.registry,omitempty"`
	DataPlane_image_repository                   string            `helm:"dataPlane.image.repository"`
	DataPlane_image_tag                          string            `helm:"dataPlane.image.tag"`
	DataPlane_initImage_registry                 string            `helm:"dataPlane.initImage.registry,omitempty"`
	DataPlane_initImage_repository               string            `helm:"dataPlane.initImage.repository"`
	DataPlane_initImage_tag                      string            `helm:"dataPlane.initImage.tag"`
	ControlPlane_kdsGlobalAddress                string            `helm:"controlPlane.kdsGlobalAddress"`
	Cni_enabled                                  bool              `helm:"cni.enabled"`
	Cni_chained                                  bool              `helm:"cni.chained"`
	Cni_net_dir                                  string            `helm:"cni.netDir"`
	Cni_bin_dir                                  string            `helm:"cni.binDir"`
	Cni_conf_name                                string            `helm:"cni.confName"`
	Cni_image_registry                           string            `helm:"cni.image.registry"`
	Cni_image_repository                         string            `helm:"cni.image.repository"`
	Cni_image_tag                                string            `helm:"cni.image.tag"`
	Cni_nodeSelector                             map[string]string `helm:"cni.nodeSelector"`
	ControlPlane_mode                            string            `helm:"controlPlane.mode"`
	ControlPlane_zone                            string            `helm:"controlPlane.zone"`
	ControlPlane_globalZoneSyncService_type      string            `helm:"controlPlane.globalZoneSyncService.type"`
	Image_registry                               string            `helm:"global.image.registry"`
	Ingress_enabled                              bool              `helm:"ingress.enabled"`
	Ingress_mesh                                 string            `helm:"ingress.mesh"`
	Ingress_drainTime                            string            `helm:"ingress.drainTime"`
	Ingress_service_type                         string            `helm:"ingress.service.type"`
	Ingress_nodeSelector                         map[string]string `helm:"ingress.nodeSelector"`
	Egress_enabled                               bool              `helm:"egress.enabled"`
	Egress_drainTime                             string            `helm:"egress.drainTime"`
	Egress_service_type                          string            `helm:"egress.service.type"`
	Egress_nodeSelector                          map[string]string `helm:"egress.nodeSelector"`
	Hooks_nodeSelector                           map[string]string `helm:"hooks.nodeSelector"`
	WithoutKubernetesConnection                  bool              // there is no HELM equivalent, HELM always require connection to Kubernetes
	ExperimentalGatewayAPI                       bool              `helm:"experimental.gatewayAPI"`
	ValueFiles                                   []string
	Values                                       []string
	SkipKinds                                    []string
	// APIVersions is a hidden, internal option
	APIVersions        []string
	DumpValues         bool
	UseNodePort        bool
	IngressUseNodePort bool
	Ebpf_bpffs_path    string
}

type ImageEnvSecret struct {
	Env    string
	Secret string
	Key    string
}

type InstallCpContext struct {
	Args                   InstallControlPlaneArgs
	InstallCpTemplateFiles func(*InstallControlPlaneArgs) (data.FileList, error)
	// When Kuma chart is embedded into other chart all the values need to have a prefix. You can set this prefix with this var.
	HELMValuesPrefix string
}

func DefaultInstallCpContext() InstallCpContext {
	return InstallCpContext{
		Args: InstallControlPlaneArgs{
			Namespace:                               "kuma-system",
			ControlPlane_image_pullPolicy:           "IfNotPresent",
			ControlPlane_image_registry:             "",
			ControlPlane_image_repository:           "kuma-cp",
			ControlPlane_image_tag:                  kuma_version.Build.Version,
			ControlPlane_service_name:               "kuma-control-plane",
			ControlPlane_envVars:                    map[string]string{},
			ControlPlane_injectorFailurePolicy:      "Fail",
			DataPlane_image_registry:                "",
			DataPlane_image_repository:              "kuma-dp",
			DataPlane_image_tag:                     kuma_version.Build.Version,
			DataPlane_initImage_registry:            "",
			DataPlane_initImage_repository:          "kuma-init",
			DataPlane_initImage_tag:                 kuma_version.Build.Version,
			Cni_enabled:                             false,
			Cni_chained:                             false,
			Cni_net_dir:                             "/etc/cni/multus/net.d",
			Cni_bin_dir:                             "/var/lib/cni/bin",
			Cni_conf_name:                           "kuma-cni.conf",
			Cni_image_registry:                      "",
			Cni_image_repository:                    "kuma-cni",
			Cni_image_tag:                           kuma_version.Build.Version,
			ControlPlane_mode:                       core.Zone,
			ControlPlane_zone:                       "",
			ControlPlane_globalZoneSyncService_type: "LoadBalancer",
			Image_registry:                          "docker.io/kumahq",
			Ingress_enabled:                         false,
			Ingress_mesh:                            "default",
			Ingress_drainTime:                       "30s",
			Ingress_service_type:                    "LoadBalancer",
			Egress_enabled:                          false,
			Egress_drainTime:                        "30s",
			Egress_service_type:                     "ClusterIP",
			Ebpf_bpffs_path:                         "/sys/fs/bpf",
		},
		InstallCpTemplateFiles: func(args *InstallControlPlaneArgs) (data.FileList, error) {
			files, err := data.ReadFiles(deployments.KumaChartFS())
			if err != nil {
				return nil, err
			}
			if !args.ExperimentalGatewayAPI {
				files = files.Filter(ExcludeGatewayAPICRDs)
			}

			return files, nil
		},
		HELMValuesPrefix: "",
	}
}

func ExcludeGatewayAPICRDs(file data.File) bool {
	return file.Name != "kuma.io_meshgatewayconfigs.yaml"
}
