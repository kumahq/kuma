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
	k8s_model "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/pkg/registry"
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

		candidates := k.candidateNames(identifier, *descriptor)
		if err := k.resourceAccess.ValidateGet(
			request.Request.Context(),
			core_model.ResourceKey{Mesh: identifier.Mesh, Name: candidates[0]},
			*descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
			return
		}

		resource := descriptor.NewObject()
		var getErr error
		for _, name := range candidates {
			getErr = k.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, identifier.Mesh))
			if getErr == nil || !store.IsNotFound(getErr) {
				break
			}
		}
		if getErr != nil {
			rest_errors.HandleError(request.Request.Context(), response, getErr, "Could not retrieve a resource")
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

// candidateNames returns store names to try when resolving a KRI. Most cases
// return a single unambiguous name. The empty-zone-on-zone-CP case is
// ambiguous: the resource may be globally originated and synced down by KDS,
// or locally admitted with a missing `kuma.io/zone` label (e.g. a policy
// created on a non-federated zone CP before that label was populated). The
// hashed-name lookup runs first to preserve existing behavior, with the
// plain name as fallback.
func (k *kriEndpoint) candidateNames(id kri.Identifier, desc core_model.ResourceTypeDescriptor) []string {
	primary := k.getCoreName(id, desc)
	if k.cpMode != config_core.Global && id.Zone == "" && !id.IsLocallyOriginated(false, k.cpZone) {
		return []string{primary, k.localName(id, desc)}
	}
	return []string{primary}
}

func (k *kriEndpoint) getCoreName(id kri.Identifier, desc core_model.ResourceTypeDescriptor) string {
	if id.IsLocallyOriginated(k.cpMode == config_core.Global, k.cpZone) {
		return k.localName(id, desc)
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
	hashed := hash.HashedName(id.Mesh, id.Name, values...)
	if k.environment == config_core.UniversalEnvironment || k.isClusterScoped(desc) {
		return hashed
	}
	return hashed + "." + k.systemNamespace
}

func (k *kriEndpoint) localName(id kri.Identifier, desc core_model.ResourceTypeDescriptor) string {
	if k.environment == config_core.UniversalEnvironment || k.isClusterScoped(desc) {
		return id.Name
	}
	namespace := id.Namespace
	if namespace == "" {
		namespace = k.systemNamespace
	}
	return id.Name + "." + namespace
}

func (k *kriEndpoint) isClusterScoped(desc core_model.ResourceTypeDescriptor) bool {
	if k.environment != config_core.KubernetesEnvironment {
		return false
	}
	obj, err := k8s_registry.Global().NewObject(desc.NewObject().GetSpec())
	if err != nil {
		return false
	}
	return obj.Scope() == k8s_model.ScopeCluster
}

func getDescriptor(resourceType core_model.ResourceType) (*core_model.ResourceTypeDescriptor, error) {
	desc, err := registry.Global().DescriptorFor(resourceType)
	if err != nil {
		return nil, err
	}

	return &desc, nil
}
