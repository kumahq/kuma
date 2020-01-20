package install

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	controlplane "github.com/Kong/kuma/app/kumactl/pkg/install/k8s/control-plane"
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
		InjectorImage           string
		InjectorFailurePolicy   string
		InjectorServiceName     string
		InjectorTlsCert         string
		InjectorTlsKey          string
		DataplaneImage          string
		DataplaneInitImage      string
		SdsTlsCert              string
		SdsTlsKey               string
	}{
		Namespace:               "kuma-system",
		ImagePullPolicy:         "IfNotPresent",
		ControlPlaneVersion:     kuma_version.Build.Version,
		ControlPlaneImage:       "kong-docker-kuma-docker.bintray.io/kuma-cp",
		ControlPlaneServiceName: "kuma-control-plane",
		AdmissionServerTlsCert:  "",
		AdmissionServerTlsKey:   "",
		InjectorImage:           "kong-docker-kuma-docker.bintray.io/kuma-injector",
		InjectorFailurePolicy:   "Ignore",
		InjectorServiceName:     "kuma-injector",
		InjectorTlsCert:         "",
		InjectorTlsKey:          "",
		DataplaneImage:          "kong-docker-kuma-docker.bintray.io/kuma-dp",
		DataplaneInitImage:      "kong-docker-kuma-docker.bintray.io/kuma-init",
		SdsTlsCert:              "",
		SdsTlsKey:               "",
	}
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Kuma Control Plane on Kubernetes",
		Long:  `Install Kuma Control Plane on Kubernetes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			if args.InjectorTlsCert == "" && args.InjectorTlsKey == "" {
				fqdn := fmt.Sprintf("%s.%s.svc", args.InjectorServiceName, args.Namespace)
				// notice that Kubernetes doesn't requires DNS SAN in a X509 cert of a WebHook
				injectorCert, err := NewSelfSignedCert(fqdn, tls.ServerCertType)
				if err != nil {
					return errors.Wrapf(err, "Failed to generate TLS certificate for %q", fqdn)
				}
				args.InjectorTlsCert = string(injectorCert.CertPEM)
				args.InjectorTlsKey = string(injectorCert.KeyPEM)
			} else if args.InjectorTlsCert == "" || args.InjectorTlsKey == "" {
				return errors.Errorf("Injector: both TLS Cert and TLS Key must be provided at the same time")
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
	cmd.Flags().StringVar(&args.InjectorImage, "injector-image", args.InjectorImage, "image of the Kuma Injector component")
	cmd.Flags().StringVar(&args.InjectorFailurePolicy, "injector-failure-policy", args.InjectorFailurePolicy, "failue policy of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.InjectorServiceName, "injector-service-name", args.InjectorServiceName, "Service name of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.InjectorTlsCert, "injector-tls-cert", args.InjectorTlsCert, "TLS certificate for the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.InjectorTlsKey, "injector-tls-key", args.InjectorTlsKey, "TLS key for the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringVar(&args.DataplaneImage, "dataplane-image", args.DataplaneImage, "image of the Kuma Dataplane component")
	cmd.Flags().StringVar(&args.DataplaneInitImage, "dataplane-init-image", args.DataplaneInitImage, "init image of the Kuma Dataplane component")
	cmd.Flags().StringVar(&args.SdsTlsCert, "sds-tls-cert", args.SdsTlsCert, "TLS certificate for the SDS server")
	cmd.Flags().StringVar(&args.SdsTlsKey, "sds-tls-key", args.SdsTlsKey, "TLS key for the SDS server")
	return cmd
}

func renderFiles(templates []data.File, args interface{}, newRenderer func(data.File) (templateRenderer, error)) ([]data.File, error) {
	renderedFiles := make([]data.File, len(templates))

	for i, template := range templates {
		renderer, err := newRenderer(template)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := renderer.Execute(&buf, args); err != nil {
			return nil, err
		}
		renderedFiles[i].Data = buf.Bytes()
	}

	return renderedFiles, nil
}

type templateRenderer interface {
	Execute(w io.Writer, data interface{}) error
}

func simpleTemplateRenderer(text data.File) (templateRenderer, error) {
	tmpl, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(text.Data))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse k8s resource template")
	}
	return tmpl, nil
}
