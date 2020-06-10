package context

import (
	"fmt"
	"io/ioutil"
	"net/url"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
)

type Context struct {
	ControlPlane *ControlPlaneContext
	Mesh         MeshContext
}

type ControlPlaneContext struct {
	SdsLocation string
	SdsTlsCert  []byte
}

type MeshContext struct {
	Resource   *mesh_core.MeshResource
	Dataplanes *mesh_core.DataplaneResourceList
}

func BuildControlPlaneContext(config kuma_cp.Config) (*ControlPlaneContext, error) {
	var cert []byte
	if config.SdsServer.TlsCertFile != "" {
		c, err := ioutil.ReadFile(config.SdsServer.TlsCertFile)
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
	if len(sdsLocation) == 0 {
		sdsLocation = fmt.Sprintf("%s:%d", config.BootstrapServer.Params.XdsHost, config.SdsServer.GrpcPort)
	}

	return &ControlPlaneContext{
		SdsLocation: sdsLocation,
		SdsTlsCert:  cert,
	}, nil
}
