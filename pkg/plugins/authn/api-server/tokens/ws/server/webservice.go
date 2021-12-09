package server

import (
	"net/http"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/access"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws"
)

var log = core.Log.WithName("user-token-ws")

type userTokenWebService struct {
	issuer issuer.UserTokenIssuer
	access access.GenerateUserTokenAccess
}

func NewWebService(issuer issuer.UserTokenIssuer, access access.GenerateUserTokenAccess) *restful.WebService {
	webservice := userTokenWebService{
		issuer: issuer,
		access: access,
	}
	return webservice.createWs()
}

func (d *userTokenWebService) createWs() *restful.WebService {
	webservice := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	webservice.Path("/tokens/user").
		Route(webservice.POST("").To(d.handleIdentityRequest))
	return webservice
}

func (d *userTokenWebService) handleIdentityRequest(request *restful.Request, response *restful.Response) {
	if err := d.access.ValidateGenerate(user.FromCtx(request.Request.Context())); err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	idReq := ws.UserTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	verr := validators.ValidationError{}
	if idReq.Name == "" {
		verr.AddViolation("name", "cannot be empty")
	}

	var validFor time.Duration
	if idReq.ValidFor == "" {
		verr.AddViolation("name", "cannot be empty")
	} else {
		dur, err := time.ParseDuration(idReq.ValidFor)
		if err != nil {
			verr.AddViolation("validFor", "is invalid: "+err.Error())
		}
		validFor = dur
	}

	if verr.HasViolations() {
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	token, err := d.issuer.Generate(request.Request.Context(), user.User{
		Name:   idReq.Name,
		Groups: idReq.Groups,
	}, validFor)
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could write a response")
	}
}
