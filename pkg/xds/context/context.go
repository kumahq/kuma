package context

import (
	"fmt"
	"io/ioutil"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
)

type Context struct {
	ControlPlane *ControlPlaneContext
	Mesh         MeshContext
}

type ControlPlaneContext struct {
	SdsLocation string
	SdsTlsCert  []byte

	DataplaneTokenFile string
}

type MeshContext struct {
	LoggingEnabled bool
	LoggingPath string
	TlsEnabled bool
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
	sdsLocation := fmt.Sprintf("%s:%d", config.BootstrapServer.Params.XdsHost, config.SdsServer.GrpcPort)
	dataplaneTokenFile := ""
	if config.Environment == kuma_cp.KubernetesEnvironment {
		dataplaneTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}

	return &ControlPlaneContext{
		SdsLocation: sdsLocation,
		SdsTlsCert:  cert,

		DataplaneTokenFile: dataplaneTokenFile,
	}, nil
}
