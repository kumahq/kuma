package test

import (
	"context"

	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_pkg_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type MockAPIServerClient struct {
	Version types.IndexResponse
}

func (m *MockAPIServerClient) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	return &m.Version, nil
}

func GetMockNewAPIServerClient() func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ApiServerClient, error) {
	return func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ApiServerClient, error) {
		return &MockAPIServerClient{
			Version: types.IndexResponse{
				Hostname: "localhost",
				Tagline:  kuma_version.Product,
				Version:  kuma_version.Build.Version,
			},
		}, nil
	}
}

// DefaultTestingRootCmd returns the DefaultRootCmd with server API mocked to return
// current version. Useful for tests which don't actually require the server but need to
// avoid extraneous warnings.
func DefaultTestingRootCmd() *cobra.Command {
	ctx := kumactl_pkg_cmd.DefaultRootContext()
	ctx.Runtime.NewAPIServerClient = GetMockNewAPIServerClient()
	return kumactl_cmd.NewRootCmd(ctx)
}
