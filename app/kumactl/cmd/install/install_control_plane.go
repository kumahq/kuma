package install

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"

	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config/core"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
)

type componentVersion struct {
	args    *install_context.InstallControlPlaneArgs
	version string
}

func (cv *componentVersion) String() string {
	return cv.version
}

func (cv *componentVersion) Set(v string) error {
	cv.version = v
	cv.args.Cni_image_tag = v
	cv.args.ControlPlane_image_tag = v
	cv.args.DataPlane_image_tag = v
	cv.args.DataPlane_initImage_tag = v
	return nil
}

func (cv *componentVersion) Type() string {
	return "string"
}

func newInstallControlPlaneCmd(ctx *install_context.InstallCpContext) *cobra.Command {
	args := ctx.Args
	useNodePort := false
	ingressUseNodePort := false
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Kuma Control Plane on Kubernetes",
		Long: `Install Kuma Control Plane on Kubernetes in its own namespace.
This command requires that the KUBECONFIG environment is set`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateArgs(args); err != nil {
				return err
			}

			if args.ExperimentalMeshGateway {
				register.RegisterGatewayTypes()
				mesh_k8s.RegisterK8SGatewayTypes()
			}

			if useNodePort && args.ControlPlane_mode == core.Global {
				args.ControlPlane_globalZoneSyncService_type = "NodePort"
			}

			if ingressUseNodePort {
				args.Ingress_service_type = "NodePort"
			}

			var kubeClientConfig *rest.Config
			if !args.WithoutKubernetesConnection {
				var err error
				kubeClientConfig, err = k8s.DefaultClientConfig("", "")
				if err != nil {
					return errors.Wrap(err, "could not detect Kubernetes configuration")
				}
			}

			templateFiles, err := ctx.InstallCpTemplateFiles(&args)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderHelmFiles(templateFiles, args, args.Namespace, ctx.HELMValuesPrefix, kubeClientConfig)
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
	componentVersion := componentVersion{
		args: &args,
	}
	// flags
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install Kuma Control Plane to")

	cmd.Flags().Var(&componentVersion, "version", "version of Kuma Control Plane components")

	cmd.Flags().StringVar(&args.ControlPlane_image_pullPolicy, "image-pull-policy", args.ControlPlane_image_pullPolicy, "image pull policy that applies to all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_image_registry, "control-plane-registry", args.ControlPlane_image_registry, "registry for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_repository, "control-plane-repository", args.ControlPlane_image_repository, "repository for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_tag, "control-plane-version", args.ControlPlane_image_tag, "version of the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_service_name, "control-plane-service-name", args.ControlPlane_service_name, "Service name of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_secret, "tls-general-secret", args.ControlPlane_tls_general_secret, "Secret that contains tls.crt, tls.key [and ca.crt when no --tls-general-ca-secret specified] for protecting Kuma in-cluster communication")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_ca_secret, "tls-general-ca-secret", args.ControlPlane_tls_general_ca_secret, "Secret that contains ca.crt that was used to sign cert for protecting Kuma in-cluster communication (ca.crt present in this secret have precedence over the one provided in --tls-general-secret)")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_caBundle, "tls-general-ca-bundle", args.ControlPlane_tls_general_secret, "Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt)")
	cmd.Flags().StringVar(&args.ControlPlane_tls_apiServer_secret, "tls-api-server-secret", args.ControlPlane_tls_apiServer_secret, "Secret that contains tls.crt, tls.key for protecting Kuma API on HTTPS")
	cmd.Flags().StringVar(&args.ControlPlane_tls_apiServer_clientCertsSecret, "tls-api-server-client-certs-secret", args.ControlPlane_tls_apiServer_clientCertsSecret, "Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS")
	cmd.Flags().StringVar(&args.ControlPlane_tls_kdsGlobalServer_secret, "tls-kds-global-server-secret", args.ControlPlane_tls_kdsGlobalServer_secret, "Secret that contains tls.crt, tls.key for protecting cross cluster communication")
	cmd.Flags().StringVar(&args.ControlPlane_tls_kdsZoneClient_secret, "tls-kds-zone-client-secret", args.ControlPlane_tls_kdsZoneClient_secret, "Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification")
	cmd.Flags().StringVar(&args.ControlPlane_injectorFailurePolicy, "injector-failure-policy", args.ControlPlane_injectorFailurePolicy, "failure policy of the mutating web hook implemented by the Kuma Injector component")
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
	cmd.Flags().StringVar(&args.ControlPlane_mode, "mode", args.ControlPlane_mode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "zone", "global"))
	cmd.Flags().StringVar(&args.ControlPlane_zone, "zone", args.ControlPlane_zone, "set the Kuma zone name")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	cmd.Flags().BoolVar(&args.Ingress_enabled, "ingress-enabled", args.Ingress_enabled, "install Kuma with an Ingress deployment, using the Data Plane image")
	cmd.Flags().StringVar(&args.Ingress_drainTime, "ingress-drain-time", args.Ingress_drainTime, "drain time for Envoy proxy")
	cmd.Flags().BoolVar(&ingressUseNodePort, "ingress-use-node-port", false, "use NodePort instead of LoadBalancer for the Ingress Service")
	cmd.Flags().BoolVar(&args.Egress_enabled, "egress-enabled", args.Egress_enabled, "install Kuma with an Egress deployment, using the Data Plane image")
	cmd.Flags().StringVar(&args.Egress_drainTime, "egress-drain-time", args.Egress_drainTime, "drain time for Envoy proxy")
	cmd.Flags().StringVar(&args.Egress_service_type, "egress-service-type", "ClusterIP", "the type for the Egress Service (ie. ClusterIP, NodePort, LoadBalancer)")
	cmd.Flags().BoolVar(&args.WithoutKubernetesConnection, "without-kubernetes-connection", false, "install without connection to Kubernetes cluster. This can be used for initial Kuma installation, but not for upgrades")
	cmd.Flags().BoolVar(&args.ExperimentalMeshGateway, "experimental-meshgateway", false, "install experimental built-in MeshGateway support")
	return cmd
}

func validateArgs(args install_context.InstallControlPlaneArgs) error {
	if err := core.ValidateCpMode(args.ControlPlane_mode); err != nil {
		return err
	}
	if args.ControlPlane_mode == core.Zone && args.ControlPlane_zone == "" {
		return errors.New("--zone is mandatory with `zone` mode")
	}
	if args.ControlPlane_mode == core.Zone && args.ControlPlane_kdsGlobalAddress == "" {
		return errors.New("--kds-global-address is mandatory with `zone` mode")
	}
	if args.ControlPlane_kdsGlobalAddress != "" {
		if args.ControlPlane_mode != core.Zone {
			return errors.New("--kds-global-address can only be used when --mode=zone")
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
