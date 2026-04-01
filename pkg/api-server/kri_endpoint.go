package api_server

import (
	"github.com/emicklei/go-restful/v3"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	"github.com/kumahq/kuma/v2/pkg/core/resources/access"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v2/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	"github.com/kumahq/kuma/v2/pkg/kds/hash"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
)

type kriEndpoint struct {
	k8sMapper       k8s.ResourceMapperFunc
	resManager      manager.ResourceManager
	resourceAccess  access.ResourceAccess
	cpMode          config_core.CpMode
	environment     config_core.EnvironmentType
	cpZone          string
	systemNamespace string
}

func (k *kriEndpoint) addFindByKriEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/_kri/{kri}").To(k.findByKriRoute()).Doc("Returns a resource by KRI").
		Param(ws.PathParameter("kri", "KRI of the resource").DataType("string")).
		Returns(200, "OK", nil).
		Returns(400, "Bad request", nil).
		Returns(404, "Not found", nil))
}

func (k *kriEndpoint) findByKriRoute() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		kriParam := request.PathParameter("kri")
		identifier, err := kri.FromString(kriParam)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, rest_errors.NewBadRequestError(err.Error()), "Could not parse KRI")
			return
		}

		descriptor, err := getDescriptor(identifier.ResourceType)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
			return
		}

		name := k.getCoreName(identifier)
		if err := k.resourceAccess.ValidateGet(
			request.Request.Context(),
			core_model.ResourceKey{Mesh: identifier.Mesh, Name: name},
			*descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
			return
		}

		resource := descriptor.NewObject()
		if err := k.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, identifier.Mesh)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
			return
		}

		res, err := formatResource(resource, request.QueryParameter("format"), k.k8sMapper, identifier.Namespace)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not format a resource")
			return
		}

		if err := response.WriteAsJson(res); err != nil {
			log.Error(err, "Could not write the find response")
			return
		}
	}
}

func (k *kriEndpoint) getCoreName(id kri.Identifier) string {
	namespace := id.Namespace
	if id.Namespace == "" {
		namespace = k.systemNamespace
	}
	if id.IsLocallyOriginated(k.cpMode == config_core.Global, k.cpZone) {
		if k.environment == config_core.UniversalEnvironment {
			return id.Name
		}
		return id.Name + "." + namespace
	}
	// Match HashSuffixMapper in pkg/kds/context/context.go: only include
	// non-empty label values so the hash is identical to what KDS produces.
	var values []string
	if id.Zone != "" {
		values = append(values, id.Zone)
	}
	if id.Namespace != "" {
		values = append(values, id.Namespace)
	}
	hashedName := hash.HashedName(id.Mesh, id.Name, values...)
	if k.environment == config_core.UniversalEnvironment {
		return hashedName
	}
	return hashedName + "." + k.systemNamespace
}

func getDescriptor(resourceType core_model.ResourceType) (*core_model.ResourceTypeDescriptor, error) {
	desc, err := registry.Global().DescriptorFor(resourceType)
	if err != nil {
		return nil, err
	}

	return &desc, nil
}
