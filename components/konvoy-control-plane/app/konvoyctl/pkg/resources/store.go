package resources

import (
	konvoy_rest "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/definitions"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	remote_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/remote"
)

func NewResourceStore(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
	client, err := apiServerClient(coordinates.Url)
	if err != nil {
		return nil, err
	}
	return remote_resources.NewStore(client, konvoy_rest.AllApis()), nil
}
