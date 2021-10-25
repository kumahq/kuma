package server

import (
	"net/http"

	"github.com/emicklei/go-restful"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/rbac"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

var log = core.Log.WithName("dataplane-token-ws")

type tokenWebService struct {
	issuer            issuer.DataplaneTokenIssuer
	zoneIngressIssuer zoneingress.TokenIssuer
	access            rbac.GenerateDataplaneTokenAccess
}

func NewWebservice(
	issuer issuer.DataplaneTokenIssuer,
	zoneIngressIssuer zoneingress.TokenIssuer,
	access rbac.GenerateDataplaneTokenAccess,
) *restful.WebService {
	ws := tokenWebService{
		issuer:            issuer,
		zoneIngressIssuer: zoneIngressIssuer,
		access:            access,
	}
	return ws.createWs()
}

func (d *tokenWebService) createWs() *restful.WebService {
	ws := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Path("/tokens").
		Route(ws.POST("").To(d.handleIdentityRequest)). // backwards compatibility
		Route(ws.POST("/dataplane").To(d.handleIdentityRequest)).
		Route(ws.POST("/zone-ingress").To(d.handleZoneIngressIdentityRequest))
	return ws
}

func (d *tokenWebService) handleIdentityRequest(request *restful.Request, response *restful.Response) {
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

	if err := d.access.ValidateGenerate(
		idReq.Name,
		idReq.Mesh,
		idReq.Tags,
		idReq.Type,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	token, err := d.issuer.Generate(issuer.DataplaneIdentity{
		Mesh: idReq.Mesh,
		Name: idReq.Name,
		Type: mesh_proto.ProxyType(idReq.Type),
		Tags: mesh_proto.MultiValueTagSetFrom(idReq.Tags),
	})
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could not write a response")
	}
}

func (d *tokenWebService) handleZoneIngressIdentityRequest(request *restful.Request, response *restful.Response) {
	idReq := types.ZoneIngressTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := d.zoneIngressIssuer.Generate(zoneingress.Identity{
		Zone: idReq.Zone,
	})
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could not write a response")
	}
}
