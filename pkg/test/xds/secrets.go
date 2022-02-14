package xds

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

var TestSecretsInfo = &secrets.Info{
	Expiration: time.Unix(2, 2),
	Generation: time.Unix(1, 1),
	Tags: map[string]map[string]bool{
		"kuma.io/service": {
			"web": true,
		},
	},
	MTLS: &mesh_proto.Mesh_Mtls{
		EnabledBackend: "ca-1",
		Backends:       nil,
	},
	IssuedBackend:     "ca-1",
	SupportedBackends: []string{"ca-1"},
}

type TestSecrets struct {
}

func get() (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	identitySecret := &core_xds.IdentitySecret{
		PemCerts: [][]byte{
			[]byte("CERT"),
		},
		PemKey: []byte("KEY"),
	}

	ca := &core_xds.CaSecret{
		PemCerts: [][]byte{
			[]byte("CA"),
		},
	}

	return identitySecret, ca, nil
}

func (*TestSecrets) GetForZoneEgress(
	*core_mesh.ZoneEgressResource,
	*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	return get()
}

func (*TestSecrets) GetForDataPlane(
	*core_mesh.DataplaneResource,
	*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	return get()
}

func (*TestSecrets) Info(model.ResourceKey) *secrets.Info {
	return TestSecretsInfo
}

func (*TestSecrets) Cleanup(model.ResourceKey) {
}

var _ secrets.Secrets = &TestSecrets{}
