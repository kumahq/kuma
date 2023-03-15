package api_server

import (
	"os"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/api-server/types"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

func addIndexWsEndpoints(ws *restful.WebService, getInstanceId func() string, getClusterId func() string, enableGUI bool, guiURL string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	var instanceId string
	var clusterId string
	if err != nil {
		return err
	}
	healthHandler := func(req *restful.Request, resp *restful.Response) {
		if instanceId == "" {
			instanceId = getInstanceId()
		}

		if clusterId == "" {
			clusterId = getClusterId()
		}

		if !enableGUI {
			guiURL = ""
		}

		response := types.IndexResponse{
			Hostname:    hostname,
			Tagline:     kuma_version.Product,
			Product:     kuma_version.Product,
			Version:     kuma_version.Build.Version,
			BasedOnKuma: kuma_version.Build.BasedOnKuma,
			InstanceId:  instanceId,
			ClusterId:   clusterId,
			GuiURL:      guiURL,
		}

		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}
	ws.Route(ws.GET("/").To(healthHandler))
	ws.Route(ws.GET("/health").To(healthHandler))
	return nil
}
