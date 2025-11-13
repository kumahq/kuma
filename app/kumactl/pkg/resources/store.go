package resources

import (
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	remote_resources "github.com/kumahq/kuma/v2/pkg/plugins/resources/remote"
	util_http "github.com/kumahq/kuma/v2/pkg/util/http"
)

func NewResourceStore(client util_http.Client, defs []core_model.ResourceTypeDescriptor) core_store.ResourceStore {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range defs {
		mapping[ws.Name] = core_rest.NewResourceApi(ws.Scope, ws.WsPath)
	}
	return remote_resources.NewStore(client, &core_rest.ApiDescriptor{
		Resources: mapping,
	})
}
