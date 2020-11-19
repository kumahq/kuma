package api_server

import (
	"github.com/emicklei/go-restful"
)

type customWsList struct {
	list []*restful.WebService
}

var GlobalCustomWsList *customWsList

func init() {
	GlobalCustomWsList = NewCustomWsList()
}

func NewCustomWsList() *customWsList {
	return &customWsList{
		list: []*restful.WebService{},
	}
}

func (c *customWsList) Add(ws *restful.WebService) {
	c.list = append(c.list, ws)
}

func (c *customWsList) AddRouteFunction(path, subpath string, handler restful.RouteFunction) {
	ws := new(restful.WebService).Path(path)
	ws.Route(ws.GET(subpath).To(handler))
	c.Add(ws)
}

func (c *customWsList) InstallAll(container *restful.Container) {
	for _, ws := range c.list {
		container.Add(ws)
	}
}
