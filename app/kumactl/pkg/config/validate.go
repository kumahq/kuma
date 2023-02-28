package config

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/client"
	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	kumactl_config "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/version"
)

func ValidateCpCoordinates(cp *kumactl_config.ControlPlane, timeout time.Duration) error {
	cl, err := client.ApiServerClient(cp.Coordinates.ApiServer, timeout)
	if err != nil {
		return err
	}
	apiServerClient := kumactl_resources.NewAPIServerClient(cl)
	response, err := apiServerClient.GetVersion(context.Background())
	if err != nil {
		return errors.Wrap(err, "could not connect to the Control Plane API Server")
	}
	if response.Tagline != version.Product && !version.IsPreviewVersion(response.Version) {
		return errors.Errorf("this CLI is for %s but the control plane you're connected to is %s. Please use the CLI for your control plane", version.Product, response.Tagline)
	}
	return nil
}
