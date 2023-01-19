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
}

type APIList struct {
	list []*restful.WebService
}

func NewAPIList() *APIList {
	return &APIList{
		list: []*restful.WebService{},
	}
}

func (c *APIList) Add(ws *restful.WebService) {
	c.list = append(c.list, ws)
}

func (c *APIList) Install(container *restful.Container) {
	for _, ws := range c.list {
		container.Add(ws)
	}
}
