package install

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	controlplane "github.com/Kong/kuma/app/kumactl/pkg/install/k8s/control-plane"
	kumacni "github.com/Kong/kuma/app/kumactl/pkg/install/k8s/kuma-cni"
	kuma_cmd "github.com/Kong/kuma/pkg/cmd"
	"github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/tls"
	kuma_version "github.com/Kong/kuma/pkg/version"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

func newInstallControlPlaneCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		Namespace               string
		ImagePullPolicy         string
		ControlPlaneVersion     string
		ControlPlaneImage       string
		ControlPlaneServiceName string
		AdmissionServerTlsCert  string
		AdmissionServerTlsKey   string
		InjectorFailurePolicy   string
		DataplaneImage          string
		DataplaneInitImage      string
		SdsTlsCert              string
		SdsTlsKey               string
		CNIEnabled              bool
		CNIImage                string
		CNIVersion              string
		KumaCpMode              string
		ClusterName             string
		GlobalRemotePortType    string
	}{
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
		ClusterName:             "",
		GlobalRemotePortType:    "LoadBalancer",
	}
	useNodePort := false
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Kuma Control Plane on Kubernetes",
		Long:  `Install Kuma Control Plane on Kubernetes in a 'kuma-system' namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := core.ValidateCpMode(args.KumaCpMode); err != nil {
				return err
			}
			if args.KumaCpMode == core.Remote && args.ClusterName == "" {
				return errors.Errorf("--cluster-name is mandatory with `remote` mode")
			}
			if useNodePort && args.KumaCpMode != core.Standalone {
				args.GlobalRemotePortType = "NodePort"
			}
			if args.AdmissionServerTlsCert == "" && args.AdmissionServerTlsKey == "" {
				fqdn := fmt.Sprintf("%s.%s.svc", args.ControlPlaneServiceName, args.Namespace)
				// notice that Kubernetes doesn't requires DNS SAN in a X509 cert of a WebHook
				admissionCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType)
				if err != nil {
					return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
				}
				args.AdmissionServerTlsCert = string(admissionCert.CertPEM)
				args.AdmissionServerTlsKey = string(admissionCert.KeyPEM)
			} else if args.AdmissionServerTlsCert == "" || args.AdmissionServerTlsKey == "" {
				return errors.Errorf("Admission Server: both TLS Cert and TLS Key must be provided at the same time")
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
			} else if args.SdsTlsCert == "" || args.SdsTlsKey == "" {
				return errors.Errorf("SDS: both TLS Cert and TLS Key must be provided at the same time")
			}

			templateFiles, err := data.ReadFiles(controlplane.Templates)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			if args.CNIEnabled {
				templateCNI, err := data.ReadFiles(kumacni.Templates)
				if err != nil {
					return errors.Wrap(err, "Failed to read template files")
				}
				templateFiles = append(templateFiles, templateCNI...)
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
	cmd.Flags().BoolVar(&args.CNIEnabled, "cni-enabled", args.CNIEnabled, "install Kuma with CNI instead of proxy init container")
	cmd.Flags().StringVar(&args.CNIImage, "cni-image", args.CNIImage, "image of Kuma CNI component, if CNIEnabled equals true")
	cmd.Flags().StringVar(&args.CNIVersion, "cni-version", args.CNIVersion, "version of the CNIImage")
	cmd.Flags().StringVar(&args.KumaCpMode, "mode", args.KumaCpMode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "remote", "global"))
	cmd.Flags().StringVar(&args.ClusterName, "cluster-name", args.ClusterName, "set the Kuma cluster name")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	return cmd
}
