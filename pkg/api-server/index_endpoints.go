package api_server

import (
	"os"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/pkg/api-server/authn"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

func addIndexWsEndpoints(ws *restful.WebService, getInstanceId, getClusterId func() string, guiURL string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	var instanceId string
	var clusterId string
	ws.Route(ws.GET("/").
		Metadata(authn.MetadataAuthKey, authn.MetadataAuthSkip).
		To(func(req *restful.Request, resp *restful.Response) {
			if instanceId == "" {
				instanceId = getInstanceId()
			}

			if clusterId == "" {
				clusterId = getClusterId()
			}

			response := types.IndexResponse{
				Hostname:   hostname,
				Product:    kuma_version.Product,
				Version:    kuma_version.Build.Version,
				InstanceId: instanceId,
				ClusterId:  clusterId,
				Gui:        guiURL,
			}
			if kuma_version.Build.BasedOnKuma != "" {
				response.BasedOnKuma = &kuma_version.Build.BasedOnKuma
			}

			if err := resp.WriteAsJson(response); err != nil {
				log.Error(err, "Could not write the index response")
			}
		}))
	return nil
}
