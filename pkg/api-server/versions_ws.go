package api_server

import (
	"io/ioutil"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
)

const VersionsFilePath = "resources/versions.json"

func versionsWs() (*restful.WebService, error) {
	json, err := ioutil.ReadFile(VersionsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read versions.json file")
	}

	ws := new(restful.WebService).Path("/versions")

	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		resp.AddHeader("content-type", "application/json")
		if _, err := resp.Write(json); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))

	return ws, nil
}
