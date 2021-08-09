package resources

import (
	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	remote_resources "github.com/kumahq/kuma/pkg/plugins/resources/remote"
)

func NewResourceStore(coordinates *config_proto.ControlPlaneCoordinates_ApiServer, defs []core_model.ResourceTypeDescriptor) (core_store.ResourceStore, error) {
	client, err := kumactl_client.ApiServerClient(coordinates)
	if err != nil {
		return nil, err
	}
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range defs {
		mapping[ws.Name] = core_rest.NewResourceApi(ws.Scope, ws.WsPath)
	}
	return remote_resources.NewStore(client, &core_rest.ApiDescriptor{
		Resources: mapping,
	}), nil
}
