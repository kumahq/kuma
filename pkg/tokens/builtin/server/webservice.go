package server

import (
	"net/http"
	"time"

	"github.com/emicklei/go-restful"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/tokens/builtin/access"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

var log = core.Log.WithName("token-ws")

type tokenWebService struct {
	issuer            issuer.DataplaneTokenIssuer
	zoneIngressIssuer zoneingress.TokenIssuer
	zoneIssuer        zone.TokenIssuer
	dpAccess          access.DataplaneTokenAccess
	zoneAccess        zone_access.ZoneTokenAccess
}

func NewWebservice(
	issuer issuer.DataplaneTokenIssuer,
	zoneIngressIssuer zoneingress.TokenIssuer,
	zoneIssuer zone.TokenIssuer,
	dpAccess access.DataplaneTokenAccess,
	zoneAccess zone_access.ZoneTokenAccess,
) *restful.WebService {
	ws := tokenWebService{
		issuer:            issuer,
		zoneIngressIssuer: zoneIngressIssuer,
		zoneIssuer:        zoneIssuer,
		dpAccess:          dpAccess,
		zoneAccess:        zoneAccess,
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
		Route(ws.POST("/zone-ingress").To(d.handleZoneIngressIdentityRequest)).
		Route(ws.POST("/zone").To(d.handleZoneIdentityRequest))
	return ws
}

func (d *tokenWebService) handleIdentityRequest(request *restful.Request, response *restful.Response) {
	idReq := types.DataplaneTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	verr := validators.ValidationError{}

	if idReq.Mesh == "" {
		verr.AddViolation("mesh", "cannot be empty")
	}

	validForErr, validFor := validateValidFor(idReq.ValidFor)
	verr.Add(validForErr)

	if idReq.Type != "" {
		if err := mesh_proto.ProxyType(idReq.Type).IsValid(); err != nil {
			verr.AddViolation("type", err.Error())
		}
	}

	if verr.HasViolations() {
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	if err := d.dpAccess.ValidateGenerateDataplaneToken(
		idReq.Name,
		idReq.Mesh,
		idReq.Tags,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	token, err := d.issuer.Generate(request.Request.Context(), issuer.DataplaneIdentity{
		Mesh: idReq.Mesh,
		Name: idReq.Name,
		Type: mesh_proto.ProxyType(idReq.Type),
		Tags: mesh_proto.MultiValueTagSetFrom(idReq.Tags),
	}, validFor)
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could not write a response")
	}
}

func validateValidFor(validForRequest string) (verr validators.ValidationError, validFor time.Duration) {
	if validForRequest == "" {
		validFor = time.Hour * 24 * 365 * 10 // 10 years. Backwards compatibility. In future releases we should make it required
	} else {
		dur, err := time.ParseDuration(validForRequest)
		if err != nil {
			verr.AddViolation("validFor", "is invalid: "+err.Error())
		}
		validFor = dur
	}
	return
}

func (d *tokenWebService) handleZoneIngressIdentityRequest(request *restful.Request, response *restful.Response) {
	idReq := types.ZoneIngressTokenRequest{}
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := d.dpAccess.ValidateGenerateZoneIngressToken(idReq.Zone, user.FromCtx(request.Request.Context())); err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	verr := validators.ValidationError{}
	if idReq.Zone == "" {
		verr.AddViolation("zone", "cannot be empty")
	}

	validForErr, validFor := validateValidFor(idReq.ValidFor)
	verr.Add(validForErr)

	if verr.HasViolations() {
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	token, err := d.zoneIngressIssuer.Generate(request.Request.Context(), zoneingress.Identity{
		Zone: idReq.Zone,
	}, validFor)
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could not write a response")
	}
}

func (d *tokenWebService) handleZoneIdentityRequest(request *restful.Request, response *restful.Response) {
	var idReq types.ZoneTokenRequest
	if err := request.ReadEntity(&idReq); err != nil {
		log.Error(err, "Could not read a request")
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := request.Request.Context()

	if err := d.zoneAccess.ValidateGenerateZoneToken(idReq.Zone, user.FromCtx(ctx)); err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	var verr validators.ValidationError

	if idReq.Scope == nil {
		verr.AddViolation("scope", "cannot be empty")
	}

	validForErr, validFor := validateValidFor(idReq.ValidFor)
	verr.Add(validForErr)

	if verr.HasViolations() {
		errors.HandleError(response, verr.OrNil(), "Invalid request")
		return
	}

	token, err := d.zoneIssuer.Generate(ctx, zone.Identity{
		Zone:  idReq.Zone,
		Scope: idReq.Scope,
	}, validFor)
	if err != nil {
		errors.HandleError(response, err, "Could not issue a token")
		return
	}

	response.Header().Set("content-type", "text/plain")
	if _, err := response.Write([]byte(token)); err != nil {
		log.Error(err, "Could not write a response")
	}
}
