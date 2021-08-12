package xds

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

type TestSecrets struct {
}

func (t *TestSecrets) Get(*core_mesh.DataplaneResource, *core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
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

func (t *TestSecrets) Info(dpKey model.ResourceKey) *secrets.Info {
	return nil
}

func (t *TestSecrets) Cleanup(dpKey model.ResourceKey) {
}

var _ secrets.Secrets = &TestSecrets{}
