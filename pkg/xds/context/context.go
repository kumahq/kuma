package context

import (
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/tls"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

type Context struct {
	ControlPlane     *ControlPlaneContext
	Mesh             MeshContext
	EnvoyAdminClient admin.EnvoyAdminClient
}

type ConnectionInfo struct {
	// Authority defines the URL that was used by the data plane to connect to the control plane
	Authority string
}

type ControlPlaneContext struct {
	AdminProxyKeyPair *tls.KeyPair
	CLACache          xds.CLACache
	Secrets           secrets.Secrets
}

type MeshContext struct {
	Resource   *core_mesh.MeshResource
	Dataplanes *core_mesh.DataplaneResourceList
	Hash       string
}

func BuildControlPlaneContext(
	config kuma_cp.Config,
	claCache xds.CLACache,
	secrets secrets.Secrets,
) (*ControlPlaneContext, error) {
	adminKeyPair, err := tls.NewSelfSignedCert("admin", tls.ServerCertType, "localhost")
	if err != nil {
		return nil, err
	}

	return &ControlPlaneContext{
		AdminProxyKeyPair: &adminKeyPair,
		CLACache:          claCache,
		Secrets:           secrets,
	}, nil
}
