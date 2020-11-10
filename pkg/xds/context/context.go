package context

import (
	"io/ioutil"
	"net/url"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type Context struct {
	ControlPlane   *ControlPlaneContext
	Mesh           MeshContext
	ConnectionInfo ConnectionInfo
}

type ConnectionInfo struct {
	// Authority defines the URL that was used by the data plane to connect to the control plane
	Authority string
}

type ControlPlaneContext struct {
	SdsLocation string
	SdsTlsCert  []byte
}

func (c Context) SDSLocation() string {
	if c.ControlPlane.SdsLocation != "" {
		return c.ControlPlane.SdsLocation
	}
	// SDS lives on the same server as XDS so we can use the URL that Dataplaned used to connect to XDS
	return c.ConnectionInfo.Authority
}

type MeshContext struct {
	Resource   *mesh_core.MeshResource
	Dataplanes *mesh_core.DataplaneResourceList
}

func BuildControlPlaneContext(config kuma_cp.Config) (*ControlPlaneContext, error) {
	var cert []byte
	if config.DpServer.TlsCertFile != "" {
		c, err := ioutil.ReadFile(config.DpServer.TlsCertFile)
		if err != nil {
			return nil, err
		}
		cert = c
	}
	var sdsLocation = ""
	if config.ApiServer.Catalog.Sds.Url != "" {
		u, err := url.Parse(config.ApiServer.Catalog.Sds.Url)
		if err != nil {
			return nil, err
		}
		sdsLocation = u.Host
	}

	return &ControlPlaneContext{
		SdsLocation: sdsLocation,
		SdsTlsCert:  cert,
	}, nil
}
