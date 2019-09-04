package install

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/install/data"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/install/k8s"
	controlplane "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/install/k8s/control-plane"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/tls"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

func newInstallControlPlaneCmd(pctx *konvoyctl_cmd.RootContext) *cobra.Command {
	args := struct {
		Namespace               string
		ImagePullPolicy         string
		ControlPlaneVersion     string
		ControlPlaneImage       string
		ControlPlaneServiceName string
		InjectorImage           string
		InjectorFailurePolicy   string
		InjectorServiceName     string
		InjectorTlsCert         string
		InjectorTlsKey          string
		DataplaneImage          string
		DataplaneInitImage      string
		DataplaneInitVersion    string
		SdsTlsCert              string
		SdsTlsKey               string
	}{
		Namespace:               "konvoy-system",
		ImagePullPolicy:         "IfNotPresent",
		ControlPlaneVersion:     "latest",
		ControlPlaneImage:       "konvoy/konvoy-control-plane",
		ControlPlaneServiceName: "konvoy-control-plane",
		InjectorImage:           "konvoy/konvoy-injector",
		InjectorFailurePolicy:   "Ignore",
		InjectorServiceName:     "konvoy-injector",
		InjectorTlsCert:         "",
		InjectorTlsKey:          "",
		DataplaneImage:          "konvoy/konvoy-dataplane",
		DataplaneInitImage:      "docker.io/istio/proxy_init",
		DataplaneInitVersion:    "1.1.2",
		SdsTlsCert:              "",
		SdsTlsKey:               "",
	}
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Konvoy Control Plane on Kubernetes",
		Long:  `Install Konvoy Control Plane on Kubernetes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if args.InjectorTlsCert == "" && args.InjectorTlsKey == "" {
				commonName := fmt.Sprintf("%s.%s.svc", args.InjectorServiceName, args.Namespace)
				injectorCert, err := NewSelfSignedCert(commonName)
				if err != nil {
					return errors.Wrapf(err, "Failed to generate TLS certificate for %q", commonName)
				}
				args.InjectorTlsCert = string(injectorCert.CertPEM)
				args.InjectorTlsKey = string(injectorCert.KeyPEM)
			} else if args.InjectorTlsCert == "" || args.InjectorTlsKey == "" {
				return errors.Errorf("Injector: both TLS Cert and TLS Key must be provided at the same time")
			}

			if args.SdsTlsCert == "" && args.SdsTlsKey == "" {
				commonName := fmt.Sprintf("%s.%s.svc", args.ControlPlaneServiceName, args.Namespace)
				sdsCert, err := NewSelfSignedCert(commonName)
				if err != nil {
					return errors.Wrapf(err, "Failed to generate TLS certificate for %q", commonName)
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

			if _, err := cmd.OutOrStdout().Write(singleFile); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}

			return nil
		},
	}
	// flags
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install Konvoy Control Plane to")
	cmd.Flags().StringVar(&args.ImagePullPolicy, "image-pull-policy", args.ImagePullPolicy, "image pull policy that applies to all components of the Konvoy Control Plane")
	cmd.Flags().StringVar(&args.ControlPlaneVersion, "control-plane-version", args.ControlPlaneVersion, "version shared by all components of the Konvoy Control Plane")
	cmd.Flags().StringVar(&args.ControlPlaneImage, "control-plane-image", args.ControlPlaneImage, "image of the Konvoy Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlaneServiceName, "control-plane-service-name", args.ControlPlaneServiceName, "Service name of the Konvoy Control Plane")
	cmd.Flags().StringVar(&args.InjectorImage, "injector-image", args.InjectorImage, "image of the Konvoy Injector component")
	cmd.Flags().StringVar(&args.InjectorFailurePolicy, "injector-failure-policy", args.InjectorFailurePolicy, "failue policy of the mutating web hook implemented by the Konvoy Injector component")
	cmd.Flags().StringVar(&args.InjectorServiceName, "injector-service-name", args.InjectorServiceName, "Service name of the mutating web hook implemented by the Konvoy Injector component")
	cmd.Flags().StringVar(&args.InjectorTlsCert, "injector-tls-cert", args.InjectorTlsCert, "TLS certificate for the mutating web hook implemented by the Konvoy Injector component")
	cmd.Flags().StringVar(&args.InjectorTlsKey, "injector-tls-key", args.InjectorTlsKey, "TLS key for the mutating web hook implemented by the Konvoy Injector component")
	cmd.Flags().StringVar(&args.DataplaneImage, "dataplane-image", args.DataplaneImage, "image of the Konvoy Dataplane component")
	cmd.Flags().StringVar(&args.DataplaneInitImage, "dataplane-init-image", args.DataplaneInitImage, "init image of the Konvoy Dataplane component")
	cmd.Flags().StringVar(&args.DataplaneInitVersion, "dataplane-init-version", args.DataplaneInitVersion, "version of the init image of the Konvoy Dataplane component")
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
		renderedFiles[i] = buf.Bytes()
	}

	return renderedFiles, nil
}

type templateRenderer interface {
	Execute(w io.Writer, data interface{}) error
}

func simpleTemplateRenderer(text data.File) (templateRenderer, error) {
	tmpl, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(text))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse k8s resource template")
	}
	return tmpl, nil
}
