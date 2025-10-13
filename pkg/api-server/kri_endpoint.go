package api_server

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful/v3"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

type kriEndpoint struct {
	k8sMapper   k8s.ResourceMapperFunc
	resManager  manager.ResourceManager
	cpMode      config_core.CpMode
	environment config_core.EnvironmentType
	cpZone      string
}

func (k *kriEndpoint) addFindByKriEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/_kri/{kri}").To(k.findByKriRoute()).Doc(fmt.Sprintf("Returns a resource by KRI")).
		Param(ws.PathParameter("kri", fmt.Sprintf("KRI of the resource")).DataType("string")).
		Returns(200, "OK", nil).
		Returns(400, "Bad request", nil).
		Returns(404, "Not found", nil))
}

func (k *kriEndpoint) findByKriRoute() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		kriParam := request.PathParameter("kri")
		namespace := request.PathParameter("namespace")
		identifier, err := kri.FromString(kriParam)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not parse KRI")
		}

		resource, err := k.findByKri(request.Request.Context(), identifier, namespace)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
		}

		res, err := formatResource(resource, request.QueryParameter("format"), k.k8sMapper, namespace)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not format a resource")
		}

		if err := response.WriteAsJson(res); err != nil {
			log.Error(err, "Could not write the find response")
		}
	}
}

func (k *kriEndpoint) findByKri(ctx context.Context, identifier kri.Identifier, namespace string) (core_model.Resource, error) {
	descriptor, err := getDescriptor(identifier.ResourceType)
	if err != nil {
		return nil, err
	}
	resource := descriptor.NewObject()
	name := k.getCoreName(identifier, namespace)
	meshName := identifier.Mesh

	if err := k.resManager.Get(ctx, resource, store.GetByKey(name, meshName)); err != nil {
		return nil, err
	}

	return resource, nil
}

func (k *kriEndpoint) getCoreName(kri kri.Identifier, namespace string) string {
	name := kri.Name
	if kri.IsLocallyOriginated(k.cpMode, k.cpZone) {
		if k.environment == config_core.UniversalEnvironment {
			return name
		} else {
			return name + "." + namespace
		}
	} else {
		return hash.HashedName(kri.Mesh, name, namespace, kri.Zone)
	}
}

func getDescriptor(resourceType core_model.ResourceType) (*core_model.ResourceTypeDescriptor, error) {
	desc, err := registry.Global().DescriptorFor(resourceType)
	if err != nil {
		return nil, err
	}

	return &desc, nil
}
