package resources

import (
	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	kuma_rest "github.com/kumahq/kuma/pkg/api-server/definitions"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	remote_resources "github.com/kumahq/kuma/pkg/plugins/resources/remote"
)

func NewResourceStore(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
	client, err := kumactl_client.ApiServerClient(coordinates)
	if err != nil {
		return nil, err
	}
	return remote_resources.NewStore(client, kuma_rest.AllApis()), nil
}
