package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
)

const (
	DefaultKonvoyNamespace = "konvoy-system"
)

var (
	// overridable by unit tests
	detectKubeConfig = konvoyctl_k8s.DetectKubeConfig
)

type configControlPlanesAddKubernetesContext struct {
	*configControlPlanesAddContext

	args struct {
		namespace string
	}
}

func newConfigControlPlanesAddKubernetesCmd(pctx *configControlPlanesAddContext) *cobra.Command {
	ctx := &configControlPlanesAddKubernetesContext{configControlPlanesAddContext: pctx}
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Add a Control Plane installed on Kubernetes",
		Long:  `Add a Control Plane installed on Kubernetes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns := ctx.args.namespace
			if ns == "" {
				return errors.Errorf(`flag "namespace" must have a non-empty value`)
			}
			config, err := detectKubeConfig()
			if err != nil {
				return errors.Wrapf(err, "Failed to detect current `kubectl` context")
			}
			if config.GetCurrentContext() == "" {
				return errors.Errorf("Failed to detect current `kubectl` context. It seems that you haven't used `kubectl` on this machine yet.")
			}
			client, err := config.NewClient()
			if err != nil {
				return errors.Wrapf(err, "Failed to connect to a target Kubernetes cluster (`kubectl` context %q)", config.GetCurrentContext())
			}
			installed, err := client.HasNamespace(ns)
			if err != nil {
				return errors.Wrapf(err, "Failed to determine whether a target Kubernetes cluster (`kubectl` context %q) has %q namespace", config.GetCurrentContext(), ns)
			}
			if !installed {
				return errors.Errorf("There is no Control Plane installed on a target Kubernetes cluster (`kubectl` context %q)", config.GetCurrentContext())
			}

			name := pctx.args.name

			cp := &config_proto.ControlPlane{
				Name: name,
				Coordinates: &config_proto.ControlPlaneCoordinates{
					Type: &config_proto.ControlPlaneCoordinates_Kubernetes_{
						Kubernetes: &config_proto.ControlPlaneCoordinates_Kubernetes{
							Kubeconfig: config.GetFilename(),
							Context:    config.GetCurrentContext(),
							Namespace:  ns,
						},
					},
				},
			}

			return pctx.AddControlPlane(cp)
		},
	}
	// flags
	cmd.Flags().StringVar(&ctx.args.namespace, "namespace", DefaultKonvoyNamespace, "Kubernetes namespace where the Control Plane have been installed to")
	return cmd
}
