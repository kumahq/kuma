package context

import (
	"io/ioutil"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/tls"
)

type Context struct {
	ControlPlane     *ControlPlaneContext
	Mesh             MeshContext
	ConnectionInfo   ConnectionInfo
	EnvoyAdminClient admin.EnvoyAdminClient
}

type ConnectionInfo struct {
	// Authority defines the URL that was used by the data plane to connect to the control plane
	Authority string
}

type ControlPlaneContext struct {
	SdsTlsCert        []byte
	AdminProxyKeyPair *tls.KeyPair
	CLACache          xds.CLACache
	DNSResolver       resolver.DNSResolver
}

func (c Context) SDSLocation() string {
	// SDS lives on the same server as XDS so we can use the URL that Dataplane used to connect to XDS
	return c.ConnectionInfo.Authority
}

type MeshContext struct {
	Resource   *mesh_core.MeshResource
	Dataplanes *mesh_core.DataplaneResourceList
	Hash       string
}

func BuildControlPlaneContext(config kuma_cp.Config, claCache xds.CLACache, dnsResolver resolver.DNSResolver) (*ControlPlaneContext, error) {
	var sdsCert []byte
	if config.DpServer.TlsCertFile != "" {
		c, err := ioutil.ReadFile(config.DpServer.TlsCertFile)
		if err != nil {
			return nil, err
		}
		sdsCert = c
	}

	adminKeyPair, err := tls.NewSelfSignedCert("admin", tls.ServerCertType, "localhost")
	if err != nil {
		return nil, err
	}

	return &ControlPlaneContext{
		SdsTlsCert:        sdsCert,
		AdminProxyKeyPair: &adminKeyPair,
		CLACache:          claCache,
		DNSResolver:       dnsResolver,
	}, nil
}
