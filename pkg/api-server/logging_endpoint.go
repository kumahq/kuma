package api_server

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v2/pkg/core/access"
	rest_errors "github.com/kumahq/kuma/v2/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

type loggingEndpoint struct {
	access   access.ControlPlaneMetadataAccess
	registry *kuma_log.ComponentLevelRegistry
}

type loggingResponse struct {
	Global     string            `json:"global"`
	Components map[string]string `json:"components"`
}

type setLevelRequest struct {
	Component string `json:"component"`
	Level     string `json:"level"`
}

func addLoggingEndpoints(ws *restful.WebService, accessControl access.ControlPlaneMetadataAccess) {
	e := &loggingEndpoint{
		access:   accessControl,
		registry: kuma_log.GlobalComponentLevelRegistry(),
	}

	ws.Route(ws.GET("/logging").To(e.list).
		Doc("List current log levels"))

	ws.Route(ws.PUT("/logging").To(e.setLevel).
		Doc("Set log level for a component or global"))

	ws.Route(ws.DELETE("/logging/{component}").To(e.resetLevel).
		Doc("Reset log level override for a component").
		Param(ws.PathParameter("component", "component name")))

	ws.Route(ws.DELETE("/logging").To(e.resetAll).
		Doc("Reset all log level overrides"))
}

func (e *loggingEndpoint) list(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	if err := e.access.ValidateView(ctx, user.FromCtx(ctx)); err != nil {
		rest_errors.HandleError(ctx, resp, err, "Access denied")
		return
	}

	overrides := e.registry.ListOverrides()
	components := make(map[string]string, len(overrides))
	for name, level := range overrides {
		components[name] = level.String()
	}

	if err := resp.WriteAsJson(loggingResponse{
		Global:     kuma_log.GetGlobalLogLevel().String(),
		Components: components,
	}); err != nil {
		log.Error(err, "Could not write logging response")
	}
}

func (e *loggingEndpoint) setLevel(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	if err := e.access.ValidateView(ctx, user.FromCtx(ctx)); err != nil {
		rest_errors.HandleError(ctx, resp, err, "Access denied")
		return
	}

	var r setLevelRequest
	if err := req.ReadEntity(&r); err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	level, err := kuma_log.ParseLogLevel(r.Level)
	if err != nil {
		if writeErr := resp.WriteHeaderAndJson(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		}, "application/json"); writeErr != nil {
			log.Error(writeErr, "Could not write error response")
		}
		return
	}

	if r.Component == "" {
		kuma_log.SetGlobalLogLevel(level)
	} else {
		e.registry.SetLevel(r.Component, level)
	}

	resp.WriteHeader(http.StatusOK)
}

func (e *loggingEndpoint) resetLevel(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	if err := e.access.ValidateView(ctx, user.FromCtx(ctx)); err != nil {
		rest_errors.HandleError(ctx, resp, err, "Access denied")
		return
	}

	component := req.PathParameter("component")
	e.registry.ResetLevel(component)
	resp.WriteHeader(http.StatusOK)
}

func (e *loggingEndpoint) resetAll(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	if err := e.access.ValidateView(ctx, user.FromCtx(ctx)); err != nil {
		rest_errors.HandleError(ctx, resp, err, "Access denied")
		return
	}

	e.registry.ResetAll()
	resp.WriteHeader(http.StatusOK)
}
