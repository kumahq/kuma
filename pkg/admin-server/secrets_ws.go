package admin_server

import (
	"context"
	"strings"

	"github.com/emicklei/go-restful"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_rest "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	rest_errors "github.com/Kong/kuma/pkg/core/rest/errors"
	secrets_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	core_validators "github.com/Kong/kuma/pkg/core/validators"
)

// This is inspired by resourceEndpoints struct for API Server.
// It's mostly copy paste, but it will be gone once we merge Admin Server with API Server
type secretsEndpoints struct {
	secretManager secrets_manager.SecretManager
}

func (s *secretsEndpoints) addFindEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/meshes/{mesh}/secrets/{name}").To(s.findSecret).
		Doc("Get a secret").
		Param(ws.PathParameter("name", "Name of a secret").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (s *secretsEndpoints) addListEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/meshes/{mesh}/secrets").To(s.listSecrets).
		Doc("List of secrets").
		Returns(200, "OK", nil))
	ws.Route(ws.GET("/secrets").To(s.listSecrets).
		Doc("List of secrets").
		Returns(200, "OK", nil))
}

func (s *secretsEndpoints) addCreateOrUpdateEndpoint(ws *restful.WebService) {
	ws.Route(ws.PUT("/meshes/{mesh}/secrets/{name}").To(s.createOrUpdateSecret).
		Doc("List of secrets").
		Returns(200, "OK", nil))
}

func (s *secretsEndpoints) addDeleteEndpoint(ws *restful.WebService) {
	ws.Route(ws.DELETE("/meshes/{mesh}/secrets/{name}").To(s.deleteResource).
		Doc("Deletes a secret").
		Param(ws.PathParameter("name", "Name of a secret").DataType("string")).
		Returns(200, "OK", nil))
}

func (s *secretsEndpoints) findSecret(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	secret := core_system.SecretResource{}

	err := s.secretManager.Get(request.Request.Context(), &secret, core_store.GetByKey(name, meshName))
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a resource")
	} else {
		res := core_rest.From.Resource(&secret)
		if err := response.WriteAsJson(res); err != nil {
			core.Log.Error(err, "Could not write the response")
		}
	}
}

func (s *secretsEndpoints) listSecrets(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")

	list := core_system.SecretResourceList{}
	if err := s.secretManager.List(request.Request.Context(), &list, core_store.ListByMesh(meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve resources")
	} else {
		restList := core_rest.From.ResourceList(&list)
		if err := response.WriteAsJson(restList); err != nil {
			rest_errors.HandleError(response, err, "Could not list resources")
		}
	}
}

func (s *secretsEndpoints) createOrUpdateSecret(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	resourceRes := core_rest.Resource{
		Spec: &system_proto.Secret{},
	}

	if err := request.ReadEntity(&resourceRes); err != nil {
		if strings.Contains(err.Error(), "illegal base64 data") {
			var verr core_validators.ValidationError
			verr.AddViolation("data", "has to be valid base64 string")
			rest_errors.HandleError(response, verr.OrNil(), "Could not process a resource")
			return
		}
		rest_errors.HandleError(response, err, "Could not process a resource")
		return
	}

	if err := s.validateResourceRequest(request, &resourceRes); err != nil {
		rest_errors.HandleError(response, err, "Could not process a resource")
		return
	}

	secret := &core_system.SecretResource{}
	if err := s.secretManager.Get(request.Request.Context(), secret, core_store.GetByKey(name, meshName)); err != nil {
		if core_store.IsResourceNotFound(err) {
			s.createResource(request.Request.Context(), name, meshName, resourceRes.Spec, response)
		} else {
			rest_errors.HandleError(response, err, "Could not find a resource")
		}
	} else {
		s.updateResource(request.Request.Context(), secret, resourceRes, response)
	}
}

func (s *secretsEndpoints) createResource(ctx context.Context, name string, meshName string, spec core_model.ResourceSpec, response *restful.Response) {
	res := &core_system.SecretResource{}
	_ = res.SetSpec(spec)
	if err := s.secretManager.Create(ctx, res, core_store.CreateByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (s *secretsEndpoints) updateResource(ctx context.Context, res *core_system.SecretResource, restRes core_rest.Resource, response *restful.Response) {
	_ = res.SetSpec(restRes.Spec)
	if err := s.secretManager.Update(ctx, res); err != nil {
		rest_errors.HandleError(response, err, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (s *secretsEndpoints) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	if err := s.secretManager.Delete(request.Request.Context(), &core_system.SecretResource{}, core_store.DeleteByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not delete a resource")
	}
}

func (s *secretsEndpoints) validateResourceRequest(request *restful.Request, resource *core_rest.Resource) error {
	var err core_validators.ValidationError
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")
	if name != resource.Meta.Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if meshName != resource.Meta.Mesh {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}
	if string(core_system.SecretType) != resource.Meta.Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	err.AddError("", core_mesh.ValidateMeta(name, meshName))
	return err.OrNil()
}
