package model

import "github.com/Kong/kuma/pkg/core/xds"

type DataplaneTokenRequest struct {
	Name string `json:"name"`
	Mesh string `json:"mesh"`
}

func (i DataplaneTokenRequest) ToProxyId() xds.ProxyId {
	return xds.ProxyId{
		Mesh:      i.Mesh,
		Namespace: "default",
		Name:      i.Name,
	}
}
