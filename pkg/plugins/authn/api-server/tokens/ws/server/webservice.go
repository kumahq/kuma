package server

import (
	"net/http"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws"
)

var log = core.Log.WithName("user-token-ws")

type userTokenWebService struct {
	issuer issuer.UserTokenIssuer
}

func NewWebService(issuer issuer.UserTokenIssuer) *restful.WebService {
	webservice := userTokenWebService{
		issuer: issuer,
	}
	return webservice.createWs()
}

func (d *userTokenWebService) createWs() *restful.WebService {
	webservice := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	webservice.Path("/user-tokens").
		Route(webservice.POST("").To(d.handleIdentityRequest))
	return webservice
}

func (d *userTokenWebService) handleIdentityRequest(request *restful.Request, response *restful.Response) {
	idReq := ws.UserTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	if idReq.Name == "" {
		verr := validators.ValidationError{}
		verr.AddViolation("name", "cannot be empty")
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	var validFor time.Duration
	if idReq.ValidFor != "" {
		dur, err := time.ParseDuration(idReq.ValidFor)
		if err != nil {
			verr := validators.ValidationError{}
			verr.AddViolation("validFor", "is invalid: "+err.Error())
			errors.HandleError(response, verr.OrNil(), "Invalid request")
			return
		}
		validFor = dur
	}

	token, err := d.issuer.Generate(user.User{
		Name:  idReq.Name,
		Group: idReq.Group,
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
