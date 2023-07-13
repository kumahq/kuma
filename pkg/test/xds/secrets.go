package xds

import (
	"context"
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
	OwnMesh: secrets.MeshInfo{
		MTLS: &mesh_proto.Mesh_Mtls{
			EnabledBackend: "ca-1",
			Backends:       nil,
		},
	},
	IssuedBackend:     "ca-1",
	SupportedBackends: []string{"ca-1"},
}

type TestSecrets struct {
}

func get(meshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, *core_xds.CaSecret, error) {
	identitySecret := &core_xds.IdentitySecret{
		PemCerts: [][]byte{
			[]byte("CERT"),
		},
		PemKey: []byte("KEY"),
	}

	cas := map[string]*core_xds.CaSecret{}
	for _, mesh := range meshes {
		cas[mesh.GetMeta().GetName()] = &core_xds.CaSecret{
			PemCerts: [][]byte{
				[]byte("CA"),
			},
		}
	}

	allInOne := &core_xds.CaSecret{
		PemCerts: [][]byte{
			[]byte("combined"),
		},
	}

	return identitySecret, cas, allInOne, nil
}

<<<<<<< HEAD
func (*TestSecrets) GetForZoneEgress(
=======
func (ts *TestSecrets) GetForZoneEgress(
	_ context.Context,
>>>>>>> df9c5f925 (fix(kuma-cp): pass context via snapshot reconciler to generateCerts (#7231))
	_ *core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	identity, cas, _, err := get([]*core_mesh.MeshResource{mesh})

	return identity, cas[mesh.GetMeta().GetName()], err
}

<<<<<<< HEAD
func (*TestSecrets) GetForDataPlane(
=======
func (ts *TestSecrets) GetForDataPlane(
	_ context.Context,
>>>>>>> df9c5f925 (fix(kuma-cp): pass context via snapshot reconciler to generateCerts (#7231))
	_ *core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	meshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, error) {
	identity, cas, _, err := get(append([]*core_mesh.MeshResource{mesh}, meshes...))
	return identity, cas, err
}

<<<<<<< HEAD
func (*TestSecrets) GetAllInOne(
=======
func (ts *TestSecrets) GetAllInOne(
	_ context.Context,
>>>>>>> df9c5f925 (fix(kuma-cp): pass context via snapshot reconciler to generateCerts (#7231))
	mesh *core_mesh.MeshResource,
	_ *core_mesh.DataplaneResource,
	meshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	identity, _, allInOne, err := get(append([]*core_mesh.MeshResource{mesh}, meshes...))
	return identity, allInOne, err
}

func (*TestSecrets) Info(model.ResourceKey) *secrets.Info {
	return TestSecretsInfo
}

func (*TestSecrets) Cleanup(model.ResourceKey) {
}

var _ secrets.Secrets = &TestSecrets{}
