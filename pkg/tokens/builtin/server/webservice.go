package server

import (
	"github.com/Kong/kuma/pkg/core/rest/errors"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
	"github.com/Kong/kuma/pkg/tokens/builtin/server/types"
	"github.com/emicklei/go-restful"
	"net/http"
)

type DataplaneTokenWebService struct {
	Issuer issuer.DataplaneTokenIssuer
}

func NewWebservice(issuer issuer.DataplaneTokenIssuer) *restful.WebService {
	ws := DataplaneTokenWebService{
		Issuer: issuer,
	}
	return ws.createWs()
}

func (d *DataplaneTokenWebService) createWs() *restful.WebService {
	ws := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Path("/tokens").
		Route(ws.POST("").To(d.handleIdentityRequest))
	return ws
}

func (d *DataplaneTokenWebService) handleIdentityRequest(request *restful.Request, response *restful.Response) {
	idReq := types.DataplaneTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	verr := validators.ValidationError{}
	if idReq.Name == "" {
		verr.AddViolation("name", "cannot be empty")
	}
	if idReq.Mesh == "" {
		verr.AddViolation("mesh", "cannot be empty")
	}
	if verr.HasViolations() {
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	token, err := d.Issuer.Generate(idReq.ToProxyId())
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could write a response")
	}
}
