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
	NoSecrets        bool
	GeneratedMeshCAs map[string]struct{}
}

func (ts *TestSecrets) get(meshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, *core_xds.CaSecret) {
	identitySecret := &core_xds.IdentitySecret{
		PemCerts: [][]byte{
			[]byte("CERT"),
		},
		PemKey: []byte("KEY"),
	}

	cas := map[string]*core_xds.CaSecret{}
	for _, mesh := range meshes {
		if ts.GeneratedMeshCAs == nil {
			ts.GeneratedMeshCAs = map[string]struct{}{}
		}
		ts.GeneratedMeshCAs[mesh.GetMeta().GetName()] = struct{}{}

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

	return identitySecret, cas, allInOne
}

func (ts *TestSecrets) GetForZoneEgress(
	_ context.Context,
	_ *core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	if ts.NoSecrets {
		return nil, nil, nil
	}
	identity, cas, _ := ts.get([]*core_mesh.MeshResource{mesh})
	return identity, cas[mesh.GetMeta().GetName()], nil
}

func (ts *TestSecrets) GetForDataPlane(
	_ context.Context,
	_ *core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	meshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, error) {
	if ts.NoSecrets {
		return nil, nil, nil
	}
	identity, cas, _ := ts.get(append([]*core_mesh.MeshResource{mesh}, meshes...))
	return identity, cas, nil
}

func (ts *TestSecrets) GetAllInOne(
	_ context.Context,
	mesh *core_mesh.MeshResource,
	_ *core_mesh.DataplaneResource,
	meshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	if ts.NoSecrets {
		return nil, nil, nil
	}
	identity, _, allInOne := ts.get(append([]*core_mesh.MeshResource{mesh}, meshes...))
	return identity, allInOne, nil
}

func (ts *TestSecrets) Info(mesh_proto.ProxyType, model.ResourceKey) *secrets.Info {
	if ts.NoSecrets {
		return nil
	}
	return TestSecretsInfo
}

func (*TestSecrets) Cleanup(mesh_proto.ProxyType, model.ResourceKey) {
}

var _ secrets.Secrets = &TestSecrets{}
