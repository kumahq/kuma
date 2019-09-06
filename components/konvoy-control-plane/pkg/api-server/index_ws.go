package api_server

import "github.com/emicklei/go-restful"

func indexWs() *restful.WebService {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		info := map[string]string{
			"tagline": "Kuma",
			"version": "0.1.0",
		}
		if err := resp.WriteAsJson(info); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return ws
}
