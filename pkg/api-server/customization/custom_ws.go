package customization

import (
	"github.com/emicklei/go-restful"
)

type APIInstaller interface {
	Install(container *restful.Container)
}

type APIManager interface {
	APIInstaller
	Add(ws *restful.WebService)
	AddRouteFunction(path, subpath string, handler restful.RouteFunction)
}

type APIList struct {
	list []*restful.WebService
}

func NewCustomWsList() *APIList {
	return &APIList{
		list: []*restful.WebService{},
	}
}

func (c *APIList) Add(ws *restful.WebService) {
	c.list = append(c.list, ws)
}

func (c *APIList) AddRouteFunction(path, subpath string, handler restful.RouteFunction) {
	ws := new(restful.WebService).Path(path)
	ws.Route(ws.GET(subpath).To(handler))
	c.Add(ws)
}

func (c *APIList) Install(container *restful.Container) {
	for _, ws := range c.list {
		container.Add(ws)
	}
}
