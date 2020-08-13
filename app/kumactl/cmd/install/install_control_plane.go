package install

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
	controlplane "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/control-plane"
	kumacni "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/kuma-cni"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/tls"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

type InstallControlPlaneArgs struct {
	Namespace               string
	ImagePullPolicy         string
	ControlPlaneVersion     string
	ControlPlaneImage       string
	ControlPlaneServiceName string
	ControlPlaneSecrets     []ImageEnvSecret
	AdmissionServerTlsCert  string
	AdmissionServerTlsKey   string
	InjectorFailurePolicy   string
	DataplaneImage          string
	DataplaneInitImage      string
	SdsTlsCert              string
	SdsTlsKey               string
	KdsTlsCert              string
	KdsTlsKey               string
	KdsGlobalAddress        string
	CNIEnabled              bool
	CNIImage                string
	CNIVersion              string
	KumaCpMode              string
	Zone                    string
	GlobalRemotePortType    string
}

type ImageEnvSecret struct {
	Env    string
	Secret string
	Key    string
}

var DefaultInstallControlPlaneArgs = InstallControlPlaneArgs{
	Namespace:               "kuma-system",
	ImagePullPolicy:         "IfNotPresent",
	ControlPlaneVersion:     kuma_version.Build.Version,
	ControlPlaneImage:       "kong-docker-kuma-docker.bintray.io/kuma-cp",
	ControlPlaneServiceName: "kuma-control-plane",
	AdmissionServerTlsCert:  "",
	AdmissionServerTlsKey:   "",
	InjectorFailurePolicy:   "Ignore",
	DataplaneImage:          "kong-docker-kuma-docker.bintray.io/kuma-dp",
	DataplaneInitImage:      "kong-docker-kuma-docker.bintray.io/kuma-init",
	SdsTlsCert:              "",
	SdsTlsKey:               "",
	CNIImage:                "lobkovilya/install-cni",
	CNIVersion:              "0.0.1",
	KumaCpMode:              core.Standalone,
	Zone:                    "",
	GlobalRemotePortType:    "LoadBalancer",
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

			if useNodePort && args.KumaCpMode != core.Standalone {
				args.GlobalRemotePortType = "NodePort"
			}

			if err := autogenerateCerts(&args); err != nil {
				return err
			}

			templateFiles, err := InstallCpTemplateFilesFn(args)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderFiles(templateFiles, args, simpleTemplateRenderer)
			if err != nil {
				return errors.Wrap(err, "Failed to render template files")
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
	cmd.Flags().StringVar(&args.ImagePullPolicy, "image-pull-policy", args.ImagePullPolicy, "image pull policy that applies to all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlaneVersion, "control-plane-version", args.ControlPlaneVersion, "version shared by all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlaneImage, "control-plane-image", args.ControlPlaneImage, "image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlaneServiceName, "control-plane-service-name", args.ControlPlaneServiceName, "Service name of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.AdmissionServerTlsCert, "admission-server-tls-cert", args.AdmissionServerTlsCert, "TLS certificate for the admission web hooks implemented by the Kuma Control Plane")
	cmd.Flags().StringVar(&args.AdmissionServerTlsKey, "admission-server-tls-key", args.AdmissionServerTlsKey, "TLS key for the admission web hooks implemented by the Kuma Control Plane")
	cmd.Flags().StringVar(&args.InjectorFailurePolicy, "injector-failure-policy", args.InjectorFailurePolicy, "failue policy of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.DataplaneImage, "dataplane-image", args.DataplaneImage, "image of the Kuma Dataplane component")
	cmd.Flags().StringVar(&args.DataplaneInitImage, "dataplane-init-image", args.DataplaneInitImage, "init image of the Kuma Dataplane component")
	cmd.Flags().StringVar(&args.SdsTlsCert, "sds-tls-cert", args.SdsTlsCert, "TLS certificate for the SDS server")
	cmd.Flags().StringVar(&args.SdsTlsKey, "sds-tls-key", args.SdsTlsKey, "TLS key for the SDS server")
	cmd.Flags().StringVar(&args.KdsTlsCert, "kds-tls-cert", args.KdsTlsCert, "TLS certificate for the KDS server")
	cmd.Flags().StringVar(&args.KdsTlsKey, "kds-tls-key", args.KdsTlsKey, "TLS key for the KDS server")
	cmd.Flags().StringVar(&args.KdsGlobalAddress, "kds-global-address", args.KdsGlobalAddress, "URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)")
	cmd.Flags().BoolVar(&args.CNIEnabled, "cni-enabled", args.CNIEnabled, "install Kuma with CNI instead of proxy init container")
	cmd.Flags().StringVar(&args.CNIImage, "cni-image", args.CNIImage, "image of Kuma CNI component, if CNIEnabled equals true")
	cmd.Flags().StringVar(&args.CNIVersion, "cni-version", args.CNIVersion, "version of the CNIImage")
	cmd.Flags().StringVar(&args.KumaCpMode, "mode", args.KumaCpMode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "remote", "global"))
	cmd.Flags().StringVar(&args.Zone, "zone", args.Zone, "set the Kuma zone name")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	return cmd
}

func validateArgs(args InstallControlPlaneArgs) error {
	if err := core.ValidateCpMode(args.KumaCpMode); err != nil {
		return err
	}
	if args.KumaCpMode == core.Remote && args.Zone == "" {
		return errors.Errorf("--zone is mandatory with `remote` mode")
	}
	if args.KumaCpMode == core.Remote && args.KdsGlobalAddress == "" {
		return errors.Errorf("--kds-global-address is mandatory with `remote` mode")
	}
	if args.KdsGlobalAddress != "" {
		if args.KumaCpMode != core.Remote {
			return errors.Errorf("--kds-global-address can only be used when --mode=remote")
		}
		u, err := url.Parse(args.KdsGlobalAddress)
		if err != nil {
			return errors.Errorf("--kds-global-address is not valid URL. The allowed format is grpcs://hostname:port")
		}
		if u.Scheme != "grpcs" {
			return errors.Errorf("--kds-global-address should start with grpcs://")
		}
	}
	if (args.AdmissionServerTlsCert == "") != (args.AdmissionServerTlsKey == "") {
		return errors.Errorf("both --admission-server-tls-cert and --admission-server-tls-key must be provided at the same time")
	}
	if (args.SdsTlsCert == "") != (args.SdsTlsKey == "") {
		return errors.Errorf("both --sds-tls-cert and --sds-tls-key must be provided at the same time")
	}
	if (args.KdsTlsCert == "") != (args.KdsTlsKey == "") {
		return errors.Errorf("both --kds-tls-cert and --kds-tls-key must be provided at the same time")
	}
	return nil
}

func autogenerateCerts(args *InstallControlPlaneArgs) error {
	if args.AdmissionServerTlsCert == "" && args.AdmissionServerTlsKey == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.ControlPlaneServiceName, args.Namespace)
		// notice that Kubernetes doesn't requires DNS SAN in a X509 cert of a WebHook
		admissionCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.AdmissionServerTlsCert = string(admissionCert.CertPEM)
		args.AdmissionServerTlsKey = string(admissionCert.KeyPEM)
	}

	if args.SdsTlsCert == "" && args.SdsTlsKey == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.ControlPlaneServiceName, args.Namespace)
		hosts := []string{
			fqdn,
			fmt.Sprintf("%s.%s", args.ControlPlaneServiceName, args.Namespace),
			args.ControlPlaneServiceName,
			"localhost",
		}
		// notice that Envoy's SDS client (Google gRPC) does require DNS SAN in a X509 cert of an SDS server
		sdsCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType, hosts...)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.SdsTlsCert = string(sdsCert.CertPEM)
		args.SdsTlsKey = string(sdsCert.KeyPEM)
	}

	if args.KdsTlsCert == "" && args.KdsTlsKey == "" {
		fqdn := fmt.Sprintf("%s.%s.svc", args.ControlPlaneServiceName, args.Namespace)
		hosts := []string{
			fqdn,
			"localhost",
		}
		kdsCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType, hosts...)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
		}
		args.KdsTlsCert = string(kdsCert.CertPEM)
		args.KdsTlsKey = string(kdsCert.KeyPEM)
	}
	return nil
}

func InstallCpTemplateFiles(args InstallControlPlaneArgs) (data.FileList, error) {
	templateFiles, err := data.ReadFiles(controlplane.Templates)
	if err != nil {
		return nil, err
	}
	if args.CNIEnabled {
		templateCNI, err := data.ReadFiles(kumacni.Templates)
		if err != nil {
			return nil, err
		}
		templateFiles = append(templateFiles, templateCNI...)
	}
	return templateFiles, nil
}
