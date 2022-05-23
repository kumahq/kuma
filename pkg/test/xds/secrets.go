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
	MeshInfos: []secrets.MeshInfo{{
		MTLS: &mesh_proto.Mesh_Mtls{
			EnabledBackend: "ca-1",
			Backends:       nil,
		},
	}},
	IssuedBackend:     "ca-1",
	SupportedBackends: []string{"ca-1"},
}

type TestSecrets struct {
}

func get(meshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, []secrets.MeshCa, secrets.MeshCa, error) {
	identitySecret := &core_xds.IdentitySecret{
		PemCerts: [][]byte{
			[]byte("CERT"),
		},
		PemKey: []byte("KEY"),
	}

	var cas []secrets.MeshCa
	for _, mesh := range meshes {
		cas = append(cas, secrets.MeshCa{
			Mesh: mesh.GetMeta().GetName(),
			CaSecret: &core_xds.CaSecret{
				PemCerts: [][]byte{
					[]byte("CA"),
				},
			},
		})
	}

	allInOne := secrets.MeshCa{
		CaSecret: &core_xds.CaSecret{
			PemCerts: [][]byte{
				[]byte("combined"),
			},
		},
		Mesh: "allmeshes",
	}

	return identitySecret, cas, allInOne, nil
}

func (*TestSecrets) GetForZoneEgress(
	_ *core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
) (*core_xds.IdentitySecret, []secrets.MeshCa, error) {
	identity, cas, _, err := get([]*core_mesh.MeshResource{mesh})
	return identity, cas, err
}

func (*TestSecrets) GetForDataPlane(
	_ *core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	meshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, []secrets.MeshCa, error) {
	identity, cas, _, err := get(append([]*core_mesh.MeshResource{mesh}, meshes...))
	return identity, cas, err
}

func (*TestSecrets) GetForGatewayListener(
	mesh *core_mesh.MeshResource,
	_ *core_mesh.DataplaneResource,
	meshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, secrets.MeshCa, error) {
	identity, _, allInOne, err := get(append([]*core_mesh.MeshResource{mesh}, meshes...))
	return identity, allInOne, err
}

func (*TestSecrets) Info(model.ResourceKey) *secrets.Info {
	return TestSecretsInfo
}

func (*TestSecrets) Cleanup(model.ResourceKey) {
}

var _ secrets.Secrets = &TestSecrets{}
