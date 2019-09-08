package resources

import (
	kuma_rest "github.com/Kong/kuma/pkg/api-server/definitions"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	remote_resources "github.com/Kong/kuma/pkg/plugins/resources/remote"
)

func NewResourceStore(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
	client, err := apiServerClient(coordinates.Url)
	if err != nil {
		return nil, err
	}
	return remote_resources.NewStore(client, kuma_rest.AllApis()), nil
}
