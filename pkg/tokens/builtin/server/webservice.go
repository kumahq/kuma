package server

import (
	"net/http"

	"github.com/emicklei/go-restful"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
)

var log = core.Log.WithName("dataplane-token-ws")

type dataplaneTokenWebService struct {
	issuer issuer.DataplaneTokenIssuer
}

func NewWebservice(issuer issuer.DataplaneTokenIssuer) *restful.WebService {
	ws := dataplaneTokenWebService{
		issuer: issuer,
	}
	return ws.createWs()
}

func (d *dataplaneTokenWebService) createWs() *restful.WebService {
	ws := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Path("/tokens").
		Route(ws.POST("").To(d.handleIdentityRequest))
	return ws
}

func (d *dataplaneTokenWebService) handleIdentityRequest(request *restful.Request, response *restful.Response) {
	idReq := types.DataplaneTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	if idReq.Mesh == "" {
		verr := validators.ValidationError{}
		verr.AddViolation("mesh", "cannot be empty")
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	token, err := d.issuer.Generate(issuer.DataplaneIdentity{
		Mesh: idReq.Mesh,
		Name: idReq.Name,
		Type: idReq.Type,
		Tags: mesh_proto.MultiValueTagSetFrom(idReq.Tags),
	})
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could write a response")
	}
}
