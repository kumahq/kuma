package api_server

import (
	"github.com/emicklei/go-restful"
)

var Versions = []byte(`{
  "kumaDp": {
    "1.0.0": {
      "envoy": "1.16.0"
    },
    "1.0.1": {
      "envoy": "1.16.0"
    },
    "1.0.2": {
      "envoy": "1.16.1"
    },
    "1.0.3": {
      "envoy": "1.16.1"
    },
    "1.0.4": {
      "envoy": "1.16.1"
    },
    "1.0.5": {
      "envoy": "1.16.2"
    },
    "1.0.6": {
      "envoy": "1.16.2"
    },
    "1.0.7": {
      "envoy": "1.16.2"
    },
    "1.0.8": {
      "envoy": "1.16.2"
    },
    "~1.1.0": {
      "envoy": "~1.17.0"
    },
    "~1.2.0": {
      "envoy": "~1.18.0"
    }
  }
}`)

func versionsWs() *restful.WebService {
	ws := new(restful.WebService).Path("/versions")

	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		resp.AddHeader("content-type", "application/json")
		if _, err := resp.Write(Versions); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))

	return ws
}
