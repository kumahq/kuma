package customization

import (
	"github.com/emicklei/go-restful/v3"
)

type APIInstaller interface {
	Install(container *restful.Container)
}

type APIManager interface {
	APIInstaller
	Add(ws *restful.WebService)
	AddFilter(filter restful.FilterFunction)
}

type APIList struct {
	list    []*restful.WebService
	filters []restful.FilterFunction
}

func NewAPIList() *APIList {
	return &APIList{
		list:    []*restful.WebService{},
		filters: []restful.FilterFunction{},
	}
}

func (c *APIList) Add(ws *restful.WebService) {
	c.list = append(c.list, ws)
}

func (c *APIList) AddFilter(f restful.FilterFunction) {
	c.filters = append(c.filters, f)
}

func (c *APIList) Install(container *restful.Container) {
	for _, ws := range c.list {
		container.Add(ws)
	}

	for _, f := range c.filters {
		container.Filter(f)
	}
}
