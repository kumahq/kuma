package install

import (
	"net/url"

	"k8s.io/client-go/rest"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
	kuma_version "github.com/kumahq/kuma/pkg/version"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	controlplane "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/control-plane"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/tls"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

type InstallControlPlaneArgs struct {
	Namespace                                    string            `helm:"namespace"`
	ControlPlane_image_pullPolicy                string            `helm:"controlPlane.image.pullPolicy"`
	ControlPlane_image_registry                  string            `helm:"controlPlane.image.registry"`
	ControlPlane_image_repository                string            `helm:"controlPlane.image.repository"`
	ControlPlane_image_tag                       string            `helm:"controlPlane.image.tag"`
	ControlPlane_service_name                    string            `helm:"controlPlane.service.name"`
	ControlPlane_tls_general_secret              string            `helm:"controlPlane.tls.general.secretName"`
	ControlPlane_tls_general_caBundle            string            `helm:"controlPlane.tls.general.caBundle"`
	ControlPlane_tls_apiServer_secret            string            `helm:"controlPlane.tls.apiServer.secretName"`
	ControlPlane_tls_apiServer_clientCertsSecret string            `helm:"controlPlane.tls.apiServer.clientCertsSecretName"`
	ControlPlane_tls_kdsGlobalServer_secret      string            `helm:"controlPlane.tls.kdsGlobalServer.secretName"`
	ControlPlane_tls_kdsRemoteClient_secret      string            `helm:"controlPlane.tls.kdsRemoteClient.secretName"`
	ControlPlane_injectorFailurePolicy           string            `helm:"controlPlane.injectorFailurePolicy"`
	ControlPlane_secrets                         []ImageEnvSecret  `helm:"controlPlane.secrets"`
	ControlPlane_envVars                         map[string]string `helm:"controlPlane.envVars"`
	DataPlane_image_registry                     string            `helm:"dataPlane.image.registry"`
	DataPlane_image_repository                   string            `helm:"dataPlane.image.repository"`
	DataPlane_image_tag                          string            `helm:"dataPlane.image.tag"`
	DataPlane_initImage_registry                 string            `helm:"dataPlane.initImage.registry"`
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
	ControlPlane_mode                            string            `helm:"controlPlane.mode"`
	ControlPlane_zone                            string            `helm:"controlPlane.zone"`
	ControlPlane_globalRemoteSyncService_type    string            `helm:"controlPlane.globalRemoteSyncService.type"`
	Ingress_enabled                              bool              `helm:"ingress.enabled"`
	Ingress_mesh                                 string            `helm:"ingress.mesh"`
	Ingress_drainTime                            string            `helm:"ingress.drainTime"`
	Ingress_service_type                         string            `helm:"ingress.service.type"`
	WithoutKubernetesConnection                  bool              // there is no HELM equivalent, HELM always require connection to Kubernetes
}

type ImageEnvSecret struct {
	Env    string
	Secret string
	Key    string
}

var DefaultInstallControlPlaneArgs = InstallControlPlaneArgs{
	Namespace:                                 "kuma-system",
	ControlPlane_image_pullPolicy:             "IfNotPresent",
	ControlPlane_image_registry:               "kong-docker-kuma-docker.bintray.io",
	ControlPlane_image_repository:             "kuma-cp",
	ControlPlane_image_tag:                    kuma_version.Build.Version,
	ControlPlane_service_name:                 "kuma-control-plane",
	ControlPlane_envVars:                      map[string]string{},
	ControlPlane_injectorFailurePolicy:        "Ignore",
	DataPlane_image_registry:                  "kong-docker-kuma-docker.bintray.io",
	DataPlane_image_repository:                "kuma-dp",
	DataPlane_image_tag:                       kuma_version.Build.Version,
	DataPlane_initImage_registry:              "kong-docker-kuma-docker.bintray.io",
	DataPlane_initImage_repository:            "kuma-init",
	DataPlane_initImage_tag:                   kuma_version.Build.Version,
	Cni_enabled:                               false,
	Cni_chained:                               false,
	Cni_net_dir:                               "/etc/cni/multus/net.d",
	Cni_bin_dir:                               "/var/lib/cni/bin",
	Cni_conf_name:                             "kuma-cni.conf",
	Cni_image_registry:                        "docker.io",
	Cni_image_repository:                      "lobkovilya/install-cni",
	Cni_image_tag:                             "0.0.2",
	ControlPlane_mode:                         core.Standalone,
	ControlPlane_zone:                         "",
	ControlPlane_globalRemoteSyncService_type: "LoadBalancer",
	Ingress_enabled:                           false,
	Ingress_mesh:                              "default",
	Ingress_drainTime:                         "30s",
	Ingress_service_type:                      "LoadBalancer",
}

var InstallCpTemplateFilesFn = InstallCpTemplateFiles

func newInstallControlPlaneCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := DefaultInstallControlPlaneArgs
	useNodePort := false
	ingressUseNodePort := false
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Kuma Control Plane on Kubernetes",
		Long: `Install Kuma Control Plane on Kubernetes in a 'kuma-system' namespace.
This command requires that the KUBECONFIG environment is set`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateArgs(args); err != nil {
				return err
			}

			if useNodePort && args.ControlPlane_mode == core.Global {
				args.ControlPlane_globalRemoteSyncService_type = "NodePort"
			}

			if ingressUseNodePort {
				args.Ingress_service_type = "NodePort"
			}

			var kubeClientConfig *rest.Config
			if !args.WithoutKubernetesConnection {
				var err error
				kubeClientConfig, err = k8s.DefaultClientConfig()
				if err != nil {
					return errors.Wrap(err, "could not detect Kubernetes configuration")
				}
			}

			templateFiles, err := InstallCpTemplateFilesFn(args)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderHelmFiles(templateFiles, args, kubeClientConfig)
			if err != nil {
				return errors.Wrap(err, "Failed to render helm template files")
			}

			sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
			if err != nil {
				return errors.Wrap(err, "Failed to sort resources by kind")
			}

			singleFile := data.JoinYAML(sortedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}

			return nil
		},
	}
	// flags
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install Kuma Control Plane to")
	cmd.Flags().StringVar(&args.ControlPlane_image_pullPolicy, "image-pull-policy", args.ControlPlane_image_pullPolicy, "image pull policy that applies to all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_image_registry, "control-plane-registry", args.ControlPlane_image_registry, "registry for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_repository, "control-plane-repository", args.ControlPlane_image_repository, "repository for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_tag, "control-plane-version", args.ControlPlane_image_tag, "version of the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_service_name, "control-plane-service-name", args.ControlPlane_service_name, "Service name of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_secret, "tls-general-secret", args.ControlPlane_tls_general_secret, "Secret that contains tls.crt, key.crt and ca.crt for protecting Kuma in-cluster communication")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_caBundle, "tls-general-ca-bundle", args.ControlPlane_tls_general_secret, "Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt)")
	cmd.Flags().StringVar(&args.ControlPlane_tls_apiServer_secret, "tls-api-server-secret", args.ControlPlane_tls_apiServer_secret, "Secret that contains tls.crt, key.crt for protecting Kuma API on HTTPS")
	cmd.Flags().StringVar(&args.ControlPlane_tls_apiServer_clientCertsSecret, "tls-api-server-client-certs-secret", args.ControlPlane_tls_apiServer_clientCertsSecret, "Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS")
	cmd.Flags().StringVar(&args.ControlPlane_tls_kdsGlobalServer_secret, "tls-kds-global-server-secret", args.ControlPlane_tls_kdsGlobalServer_secret, "Secret that contains tls.crt, key.crt for protecting cross cluster communication")
	cmd.Flags().StringVar(&args.ControlPlane_tls_kdsRemoteClient_secret, "tls-kds-remote-client-secret", args.ControlPlane_tls_kdsRemoteClient_secret, "Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification")
	cmd.Flags().StringVar(&args.ControlPlane_injectorFailurePolicy, "injector-failure-policy", args.ControlPlane_injectorFailurePolicy, "failue policy of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringToStringVar(&args.ControlPlane_envVars, "env-var", args.ControlPlane_envVars, "environment variables that will be passed to the control plane")
	cmd.Flags().StringVar(&args.DataPlane_image_registry, "dataplane-registry", args.DataPlane_image_registry, "registry for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_image_repository, "dataplane-repository", args.DataPlane_image_repository, "repository for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_image_tag, "dataplane-version", args.DataPlane_image_tag, "version of the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_registry, "dataplane-init-registry", args.DataPlane_initImage_registry, "registry for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_repository, "dataplane-init-repository", args.DataPlane_initImage_repository, "repository for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_tag, "dataplane-init-version", args.DataPlane_initImage_tag, "version of the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.ControlPlane_kdsGlobalAddress, "kds-global-address", args.ControlPlane_kdsGlobalAddress, "URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)")
	cmd.Flags().BoolVar(&args.Cni_enabled, "cni-enabled", args.Cni_enabled, "install Kuma with CNI instead of proxy init container")
	cmd.Flags().BoolVar(&args.Cni_chained, "cni-chained", args.Cni_chained, "enable chained CNI installation")
	cmd.Flags().StringVar(&args.Cni_net_dir, "cni-net-dir", args.Cni_net_dir, "set the CNI install directory")
	cmd.Flags().StringVar(&args.Cni_bin_dir, "cni-bin-dir", args.Cni_bin_dir, "set the CNI binary directory")
	cmd.Flags().StringVar(&args.Cni_conf_name, "cni-conf-name", args.Cni_conf_name, "set the CNI configuration name")
	cmd.Flags().StringVar(&args.Cni_image_registry, "cni-registry", args.Cni_image_registry, "registry for the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Cni_image_repository, "cni-repository", args.Cni_image_repository, "repository for the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Cni_image_tag, "cni-version", args.Cni_image_tag, "version of the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.ControlPlane_mode, "mode", args.ControlPlane_mode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "remote", "global"))
	cmd.Flags().StringVar(&args.ControlPlane_zone, "zone", args.ControlPlane_zone, "set the Kuma zone name")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	cmd.Flags().BoolVar(&args.Ingress_enabled, "ingress-enabled", args.Cni_enabled, "install Kuma with an Ingress deployment, using the Data Plane image")
	cmd.Flags().StringVar(&args.Ingress_drainTime, "ingress-drain-time", args.Ingress_drainTime, "drain time for Envoy proxy")
	cmd.Flags().BoolVar(&ingressUseNodePort, "ingress-use-node-port", false, "use NodePort instead of LoadBalancer for the Ingress Service")
	cmd.Flags().BoolVar(&args.WithoutKubernetesConnection, "without-kubernetes-connection", false, "install without connection to Kubernetes cluster. This can be used for initial Kuma installation, but not for upgrades")
	return cmd
}

func validateArgs(args InstallControlPlaneArgs) error {
	if err := core.ValidateCpMode(args.ControlPlane_mode); err != nil {
		return err
	}
	if args.ControlPlane_mode == core.Remote && args.ControlPlane_zone == "" {
		return errors.New("--zone is mandatory with `remote` mode")
	}
	if args.ControlPlane_mode == core.Remote && args.ControlPlane_kdsGlobalAddress == "" {
		return errors.New("--kds-global-address is mandatory with `remote` mode")
	}
	if args.ControlPlane_kdsGlobalAddress != "" {
		if args.ControlPlane_mode != core.Remote {
			return errors.New("--kds-global-address can only be used when --mode=remote")
		}
		u, err := url.Parse(args.ControlPlane_kdsGlobalAddress)
		if err != nil {
			return errors.New("--kds-global-address is not valid URL. The allowed format is grpcs://hostname:port")
		}
		if u.Scheme != "grpcs" {
			return errors.New("--kds-global-address should start with grpcs://")
		}
	}
	if (args.ControlPlane_tls_general_secret == "") != (args.ControlPlane_tls_general_caBundle == "") {
		return errors.New("--tls-general-secret and --tls-general-ca-bundle must be provided at the same time")
	}
	return nil
}

func InstallCpTemplateFiles(args InstallControlPlaneArgs) (data.FileList, error) {
	templateFiles, err := data.ReadFiles(controlplane.HelmTemplates)
	if err != nil {
		return nil, err
	}

	return templateFiles, nil
}
