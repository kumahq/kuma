package install

import (
	"fmt"
	"net/url"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"

	kuma_version "github.com/kumahq/kuma/pkg/version"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	controlplane "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/control-plane"
	kumacni "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/kuma-cni"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/tls"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

type InstallControlPlaneArgs struct {
	Release_Namespace                                string
	Values_controlPlane_image_pullPolicy             string
	Values_controlPlane_image_registry               string
	Values_controlPlane_image_repositry              string
	Values_controlPlane_image_tag                    string
	Values_controlPlane_service_name                 string
	Values_controlPlane_tls_admission_cert           string
	Values_controlPlane_tls_admission_key            string
	Values_controlPlane_tls_sds_cert                 string
	Values_controlPlane_tls_sds_key                  string
	Values_controlPlane_tls_kds_cert                 string
	Values_controlPlane_tls_kds_key                  string
	Values_controlPlane_injectorFailurePolicy        string
	Values_dataPlane_image_registry                  string
	Values_dataPlane_image_repositry                 string
	Values_dataPlane_image_tag                       string
	Values_dataPlane_initImage_registry              string
	Values_dataPlane_initImage_repositry             string
	Values_dataPlane_initImage_tag                   string
	Values_controlPlane_kdsGlobalAddress             string
	Values_cni_enabled                               bool
	Values_cni_image_registry                        string
	Values_cni_image_repositry                       string
	Values_cni_image_tag                             string
	Values_controlPlane_mode                         string
	Values_controlPlane_zone                         string
	Values_controlPlane_globalRemoteSyncService_type string
}

var DefaultInstallControlPlaneArgs = InstallControlPlaneArgs{
	Release_Namespace:                                "kuma-system",
	Values_controlPlane_image_pullPolicy:             "IfNotPresent",
	Values_controlPlane_image_registry:               "kong-docker-kuma-docker.bintray.io",
	Values_controlPlane_image_repositry:              "kuma-cp",
	Values_controlPlane_image_tag:                    kuma_version.Build.Version,
	Values_controlPlane_service_name:                 "kuma-control-plane",
	Values_controlPlane_tls_admission_cert:           "",
	Values_controlPlane_tls_admission_key:            "",
	Values_controlPlane_tls_sds_cert:                 "",
	Values_controlPlane_tls_sds_key:                  "",
	Values_controlPlane_tls_kds_cert:                 "",
	Values_controlPlane_tls_kds_key:                  "",
	Values_controlPlane_injectorFailurePolicy:        "Ignore",
	Values_dataPlane_image_registry:                  "kong-docker-kuma-docker.bintray.io",
	Values_dataPlane_image_repositry:                 "kuma-dp",
	Values_dataPlane_image_tag:                       kuma_version.Build.Version,
	Values_dataPlane_initImage_registry:              "kong-docker-kuma-docker.bintray.io",
	Values_dataPlane_initImage_repositry:             "kuma-init",
	Values_dataPlane_initImage_tag:                   kuma_version.Build.Version,
	Values_cni_image_registry:                        "docker.io",
	Values_cni_image_repositry:                       "lobkovilya/install-cni",
	Values_cni_image_tag:                             "0.0.1",
	Values_controlPlane_mode:                         core.Standalone,
	Values_controlPlane_zone:                         "",
	Values_controlPlane_globalRemoteSyncService_type: "LoadBalancer",
}

var InstallCpTemplateFilesFn = InstallCpTemplateFiles

func newInstallControlPlaneCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := DefaultInstallControlPlaneArgs
	useNodePort := false
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Kuma Control Plane on Kubernetes",
		Long:  `Install Kuma Control Plane on Kubernetes in a 'kuma-system' namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateArgs(args); err != nil {
				return err
			}

			if useNodePort && args.Values_controlPlane_mode == core.Global {
				args.Values_controlPlane_globalRemoteSyncService_type = "NodePort"
			}

			if err := autogenerateCerts(&args); err != nil {
				return err
			}

			templateFiles, err := InstallCpTemplateFilesFn(args)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderHelmFiles(templateFiles, args)
			if err != nil {
				return errors.Wrap(err, "Failed to render helm template files")
			}

			sortedResources := k8s.SortResourcesByKind(renderedFiles)

			singleFile := data.JoinYAML(sortedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}

			return nil
		},
	}
	// flags
	cmd.Flags().StringVar(&args.Release_Namespace, "namespace", args.Release_Namespace, "namespace to install Kuma Control Plane to")
	cmd.Flags().StringVar(&args.Values_controlPlane_image_pullPolicy, "image-pull-policy", args.Values_controlPlane_image_pullPolicy, "image pull policy that applies to all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.Values_controlPlane_image_registry, "control-plane-registry", args.Values_controlPlane_image_registry, "registry for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.Values_controlPlane_image_repositry, "control-plane-repository", args.Values_controlPlane_image_repositry, "repository for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.Values_controlPlane_image_tag, "control-plane-version", args.Values_controlPlane_image_tag, "version of the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.Values_controlPlane_service_name, "control-plane-service-name", args.Values_controlPlane_service_name, "Service name of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.Values_controlPlane_tls_admission_cert, "admission-server-tls-cert", args.Values_controlPlane_tls_admission_cert, "TLS certificate for the admission web hooks implemented by the Kuma Control Plane")
	cmd.Flags().StringVar(&args.Values_controlPlane_tls_admission_key, "admission-server-tls-key", args.Values_controlPlane_tls_admission_key, "TLS key for the admission web hooks implemented by the Kuma Control Plane")
	cmd.Flags().StringVar(&args.Values_controlPlane_tls_sds_cert, "sds-tls-cert", args.Values_controlPlane_tls_sds_cert, "TLS certificate for the SDS server")
	cmd.Flags().StringVar(&args.Values_controlPlane_tls_sds_key, "sds-tls-key", args.Values_controlPlane_tls_sds_key, "TLS key for the SDS server")
	cmd.Flags().StringVar(&args.Values_controlPlane_tls_kds_cert, "kds-tls-cert", args.Values_controlPlane_tls_kds_cert, "TLS certificate for the KDS server")
	cmd.Flags().StringVar(&args.Values_controlPlane_tls_kds_key, "kds-tls-key", args.Values_controlPlane_tls_kds_key, "TLS key for the KDS server")
	cmd.Flags().StringVar(&args.Values_controlPlane_injectorFailurePolicy, "injector-failure-policy", args.Values_controlPlane_injectorFailurePolicy, "failue policy of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.Values_dataPlane_image_registry, "dataplane-registry", args.Values_dataPlane_image_registry, "registry for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.Values_dataPlane_image_repositry, "dataplane-repository", args.Values_dataPlane_image_repositry, "repository for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.Values_dataPlane_image_tag, "dataplane-version", args.Values_dataPlane_image_tag, "version of the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.Values_dataPlane_initImage_registry, "dataplane-init-registry", args.Values_dataPlane_image_registry, "registry for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.Values_dataPlane_initImage_repositry, "dataplane-init-repository", args.Values_dataPlane_image_repositry, "repository for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.Values_dataPlane_initImage_tag, "dataplane-init-version", args.Values_dataPlane_image_tag, "version of the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.Values_controlPlane_kdsGlobalAddress, "kds-global-address", args.Values_controlPlane_kdsGlobalAddress, "URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)")
	cmd.Flags().BoolVar(&args.Values_cni_enabled, "cni-enabled", args.Values_cni_enabled, "install Kuma with CNI instead of proxy init container")
	cmd.Flags().StringVar(&args.Values_cni_image_registry, "cni-registry", args.Values_dataPlane_image_registry, "registry for the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Values_cni_image_repositry, "cni-repository", args.Values_dataPlane_image_repositry, "repository for the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Values_cni_image_tag, "cni-version", args.Values_dataPlane_image_tag, "version of the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Values_controlPlane_mode, "mode", args.Values_controlPlane_mode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "remote", "global"))
	cmd.Flags().StringVar(&args.Values_controlPlane_zone, "zone", args.Values_controlPlane_zone, "set the Kuma zone name")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	return cmd
}

func validateArgs(args InstallControlPlaneArgs) error {
	if err := core.ValidateCpMode(args.Values_controlPlane_mode); err != nil {
		return err
	}
	if args.Values_controlPlane_mode == core.Remote && args.Values_controlPlane_zone == "" {
		return errors.Errorf("--zone is mandatory with `remote` mode")
	}
	if args.Values_controlPlane_mode == core.Remote && args.Values_controlPlane_kdsGlobalAddress == "" {
		return errors.Errorf("--kds-global-address is mandatory with `remote` mode")
	}
	if args.Values_controlPlane_kdsGlobalAddress != "" {
		if args.Values_controlPlane_mode != core.Remote {
			return errors.Errorf("--kds-global-address can only be used when --mode=remote")
		}
		u, err := url.Parse(args.Values_controlPlane_kdsGlobalAddress)
		if err != nil {
			return errors.Errorf("--kds-global-address is not valid URL. The allowed format is grpcs://hostname:port")
		}
		if u.Scheme != "grpcs" {
			return errors.Errorf("--kds-global-address should start with grpcs://")
		}
	}
	if (args.Values_controlPlane_tls_admission_cert == "") != (args.Values_controlPlane_tls_admission_key == "") {
		return errors.Errorf("both --admission-server-tls-cert and --admission-server-tls-key must be provided at the same time")
	}
	if (args.Values_controlPlane_tls_sds_cert == "") != (args.Values_controlPlane_tls_sds_key == "") {
		return errors.Errorf("both --sds-tls-cert and --sds-tls-key must be provided at the same time")
	}
	if (args.Values_controlPlane_tls_kds_cert == "") != (args.Values_controlPlane_tls_kds_key == "") {
		return errors.Errorf("both --kds-tls-cert and --kds-tls-key must be provided at the same time")
	}
	return nil
}

func autogenerateCerts(args *InstallControlPlaneArgs) error {
	if args.Values_controlPlane_tls_admission_cert == "" && args.Values_controlPlane_tls_admission_key == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.Values_controlPlane_service_name, args.Release_Namespace)
		// notice that Kubernetes doesn't requires DNS SAN in a X509 cert of a WebHook
		admissionCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.Values_controlPlane_tls_admission_cert = string(admissionCert.CertPEM)
		args.Values_controlPlane_tls_admission_key = string(admissionCert.KeyPEM)
	}

	if args.Values_controlPlane_tls_sds_cert == "" && args.Values_controlPlane_tls_sds_key == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.Values_controlPlane_service_name, args.Release_Namespace)
		hosts := []string{
			fqdn,
			fmt.Sprintf("%s.%s", args.Values_controlPlane_service_name, args.Release_Namespace),
			args.Values_controlPlane_service_name,
			"localhost",
		}
		// notice that Envoy's SDS client (Google gRPC) does require DNS SAN in a X509 cert of an SDS server
		sdsCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType, hosts...)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.Values_controlPlane_tls_sds_cert = string(sdsCert.CertPEM)
		args.Values_controlPlane_tls_sds_key = string(sdsCert.KeyPEM)
	}

	if args.Values_controlPlane_tls_kds_cert == "" && args.Values_controlPlane_tls_kds_key == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.Values_controlPlane_service_name, args.Release_Namespace)
		hosts := []string{
			fqdn,
			"localhost",
		}
		kdsCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType, hosts...)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.Values_controlPlane_tls_kds_cert = string(kdsCert.CertPEM)
		args.Values_controlPlane_tls_kds_key = string(kdsCert.KeyPEM)
	}
	return nil
}

func InstallCpTemplateFiles(args InstallControlPlaneArgs) (data.FileList, error) {
	templateFiles, err := data.ReadFiles(controlplane.HelmTemplates)
	if err != nil {
		return nil, err
	}
	if args.Values_cni_enabled {
		templateCNI, err := data.ReadFiles(kumacni.Templates)
		if err != nil {
			return nil, err
		}
		templateFiles = append(templateFiles, templateCNI...)
	}
	return templateFiles, nil
}
