package context

import (
	"fmt"
	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	"io/ioutil"
)

type Context struct {
	ControlPlane *ControlPlaneContext
}

type ControlPlaneContext struct {
	SdsLocation string
	SdsTlsCert  []byte
}

func BuildControlPlaneContext(config konvoy_cp.Config) (*ControlPlaneContext, error) {
	var cert []byte
	if config.SdsServer.TlsCertFile != "" {
		c, err := ioutil.ReadFile(config.SdsServer.TlsCertFile)
		if err != nil {
			return nil, err
		}
		cert = c
	}
	sdsLocation := fmt.Sprintf("%s:%d", config.BootstrapServer.Params.XdsHost, config.SdsServer.GrpcPort)
	return &ControlPlaneContext{
		SdsLocation: sdsLocation,
		SdsTlsCert:  cert,
	}, nil
}
