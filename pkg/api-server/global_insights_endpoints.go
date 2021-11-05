package api_server

import (
	"strings"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
)

type globalInsightsEndpoints struct {
	resManager     manager.ResourceManager
	resourceAccess access.ResourceAccess
	resources      map[string]globalInsightsStat
}

func newGlobalInsightsEndpoints(
	resManager manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) *globalInsightsEndpoints {
	return &globalInsightsEndpoints{
		resManager:     resManager,
		resourceAccess: resourceAccess,
		resources:      map[string]globalInsightsStat{},
	}
}

type globalInsightsStat struct {
	Total uint32 `json:"total"`
}

type globalInsightsResponse struct {
	Type         string                        `json:"type"`
	CreationTime time.Time                     `json:"creationTime"`
	Resources    map[string]globalInsightsStat `json:"resources"`
}

func newGlobalInsightsResponse(resources map[string]globalInsightsStat) *globalInsightsResponse {
	return &globalInsightsResponse{
		Type:         "GlobalInsights",
		CreationTime: core.Now(),
		Resources:    resources,
	}
}

func (r *globalInsightsEndpoints) addEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/global-insights").To(r.inspectGlobalResources).
		Doc("Inspect all global resources").
		Returns(200, "OK", nil))
}

func (r *globalInsightsEndpoints) inspectGlobalResources(request *restful.Request, response *restful.Response) {
	for _, descriptor := range registry.Global().ObjectDescriptors() {
		if descriptor.Scope != model.ScopeGlobal ||
			descriptor.Name == system.ConfigType ||
			strings.HasSuffix(string(descriptor.Name), "Insight") {
			continue
		}

		resources := descriptor.NewList()
		if err := r.resManager.List(request.Request.Context(), resources); err != nil {
			rest_errors.HandleError(response, err, "Could not retrieve global insights")
			return
		}

		r.resources[string(descriptor.Name)] = globalInsightsStat{
			Total: uint32(len(resources.GetItems())),
		}
	}

	insights := newGlobalInsightsResponse(r.resources)

	if err := response.WriteAsJson(insights); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve global insights")
	}
}
