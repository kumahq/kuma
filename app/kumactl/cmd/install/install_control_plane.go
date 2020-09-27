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
	Namespace                                 string           `helm:"namespace"`
	ControlPlane_image_pullPolicy             string           `helm:"controlPlane.image.pullPolicy"`
	ControlPlane_image_registry               string           `helm:"controlPlane.image.registry"`
	ControlPlane_image_repositry              string           `helm:"controlPlane.image.repositry"`
	ControlPlane_image_tag                    string           `helm:"controlPlane.image.tag"`
	ControlPlane_service_name                 string           `helm:"controlPlane.service.name"`
	ControlPlane_tls_cert                     string           `helm:"controlPlane.tls.cert"`
	ControlPlane_tls_key                      string           `helm:"controlPlane.tls.key"`
	ControlPlane_injectorFailurePolicy        string           `helm:"controlPlane.injectorFailurePolicy"`
	ControlPlane_secrets                      []ImageEnvSecret `helm:"controlPlane.secrets"`
	DataPlane_image_registry                  string           `helm:"dataPlane.image.registry"`
	DataPlane_image_repositry                 string           `helm:"dataPlane.image.repositry"`
	DataPlane_image_tag                       string           `helm:"dataPlane.image.tag"`
	DataPlane_initImage_registry              string           `helm:"dataPlane.initImage.registry"`
	DataPlane_initImage_repositry             string           `helm:"dataPlane.initImage.repositry"`
	DataPlane_initImage_tag                   string           `helm:"dataPlane.initImage.tag"`
	ControlPlane_kdsGlobalAddress             string           `helm:"controlPlane.kdsGlobalAddress"`
	Cni_enabled                               bool             `helm:"cni.enabled"`
	Cni_image_registry                        string           `helm:"cni.image.registry"`
	Cni_image_repositry                       string           `helm:"cni.image.repositry"`
	Cni_image_tag                             string           `helm:"cni.image.tag"`
	ControlPlane_mode                         string           `helm:"controlPlane.mode"`
	ControlPlane_zone                         string           `helm:"controlPlane.zone"`
	ControlPlane_globalRemoteSyncService_type string           `helm:"controlPlane.globalRemoteSyncService.type"`
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
	ControlPlane_image_repositry:              "kuma-cp",
	ControlPlane_image_tag:                    kuma_version.Build.Version,
	ControlPlane_service_name:                 "kuma-control-plane",
	ControlPlane_tls_cert:                     "",
	ControlPlane_tls_key:                      "",
	ControlPlane_injectorFailurePolicy:        "Ignore",
	DataPlane_image_registry:                  "kong-docker-kuma-docker.bintray.io",
	DataPlane_image_repositry:                 "kuma-dp",
	DataPlane_image_tag:                       kuma_version.Build.Version,
	DataPlane_initImage_registry:              "kong-docker-kuma-docker.bintray.io",
	DataPlane_initImage_repositry:             "kuma-init",
	DataPlane_initImage_tag:                   kuma_version.Build.Version,
	Cni_image_registry:                        "docker.io",
	Cni_image_repositry:                       "lobkovilya/install-cni",
	Cni_image_tag:                             "0.0.1",
	ControlPlane_mode:                         core.Standalone,
	ControlPlane_zone:                         "",
	ControlPlane_globalRemoteSyncService_type: "LoadBalancer",
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

			if useNodePort && args.ControlPlane_mode == core.Global {
				args.ControlPlane_globalRemoteSyncService_type = "NodePort"
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
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install Kuma Control Plane to")
	cmd.Flags().StringVar(&args.ControlPlane_image_pullPolicy, "image-pull-policy", args.ControlPlane_image_pullPolicy, "image pull policy that applies to all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_image_registry, "control-plane-registry", args.ControlPlane_image_registry, "registry for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_repositry, "control-plane-repository", args.ControlPlane_image_repositry, "repository for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_tag, "control-plane-version", args.ControlPlane_image_tag, "version of the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_service_name, "control-plane-service-name", args.ControlPlane_service_name, "Service name of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_tls_cert, "tls-cert", args.ControlPlane_tls_cert, "TLS certificate for Kuma Control Plane servers")
	cmd.Flags().StringVar(&args.ControlPlane_tls_key, "tls-key", args.ControlPlane_tls_key, "TLS key for Kuma Control Plane servers")
	cmd.Flags().StringVar(&args.ControlPlane_injectorFailurePolicy, "injector-failure-policy", args.ControlPlane_injectorFailurePolicy, "failue policy of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.DataPlane_image_registry, "dataplane-registry", args.DataPlane_image_registry, "registry for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_image_repositry, "dataplane-repository", args.DataPlane_image_repositry, "repository for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_image_tag, "dataplane-version", args.DataPlane_image_tag, "version of the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_registry, "dataplane-init-registry", args.DataPlane_initImage_registry, "registry for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_repositry, "dataplane-init-repository", args.DataPlane_initImage_repositry, "repository for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_tag, "dataplane-init-version", args.DataPlane_initImage_tag, "version of the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.ControlPlane_kdsGlobalAddress, "kds-global-address", args.ControlPlane_kdsGlobalAddress, "URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)")
	cmd.Flags().BoolVar(&args.Cni_enabled, "cni-enabled", args.Cni_enabled, "install Kuma with CNI instead of proxy init container")
	cmd.Flags().StringVar(&args.Cni_image_registry, "cni-registry", args.Cni_image_registry, "registry for the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Cni_image_repositry, "cni-repository", args.Cni_image_repositry, "repository for the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.Cni_image_tag, "cni-version", args.Cni_image_tag, "version of the image of the Kuma CNI component")
	cmd.Flags().StringVar(&args.ControlPlane_mode, "mode", args.ControlPlane_mode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "remote", "global"))
	cmd.Flags().StringVar(&args.ControlPlane_zone, "zone", args.ControlPlane_zone, "set the Kuma zone name")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	return cmd
}

func validateArgs(args InstallControlPlaneArgs) error {
	if err := core.ValidateCpMode(args.ControlPlane_mode); err != nil {
		return err
	}
	if args.ControlPlane_mode == core.Remote && args.ControlPlane_zone == "" {
		return errors.Errorf("--zone is mandatory with `remote` mode")
	}
	if args.ControlPlane_mode == core.Remote && args.ControlPlane_kdsGlobalAddress == "" {
		return errors.Errorf("--kds-global-address is mandatory with `remote` mode")
	}
	if args.ControlPlane_kdsGlobalAddress != "" {
		if args.ControlPlane_mode != core.Remote {
			return errors.Errorf("--kds-global-address can only be used when --mode=remote")
		}
		u, err := url.Parse(args.ControlPlane_kdsGlobalAddress)
		if err != nil {
			return errors.Errorf("--kds-global-address is not valid URL. The allowed format is grpcs://hostname:port")
		}
		if u.Scheme != "grpcs" {
			return errors.Errorf("--kds-global-address should start with grpcs://")
		}
	}
	if (args.ControlPlane_tls_cert == "") != (args.ControlPlane_tls_key == "") {
		return errors.Errorf("both --tls-cert and --tls-key must be provided at the same time")
	}
	return nil
}

func autogenerateCerts(args *InstallControlPlaneArgs) error {
	if args.ControlPlane_tls_cert == "" && args.ControlPlane_tls_key == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.ControlPlane_service_name, args.Namespace)
		hosts := []string{
			fqdn,
			fmt.Sprintf("%s.%s", args.ControlPlane_service_name, args.Namespace),
			"localhost",
		}
		cert, err := NewSelfSignedCert(fqdn, tls.ServerCertType, hosts...)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.ControlPlane_tls_cert = string(cert.CertPEM)
		args.ControlPlane_tls_key = string(cert.KeyPEM)
	}
	return nil
}

func InstallCpTemplateFiles(args InstallControlPlaneArgs) (data.FileList, error) {
	templateFiles, err := data.ReadFiles(controlplane.HelmTemplates)
	if err != nil {
		return nil, err
	}
	if args.Cni_enabled {
		templateCNI, err := data.ReadFiles(kumacni.Templates)
		if err != nil {
			return nil, err
		}
		templateFiles = append(templateFiles, templateCNI...)
	}
	return templateFiles, nil
}
