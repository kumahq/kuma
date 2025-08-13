package builders

import (
	"context"

	meshtrust_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshTrustBuilder struct {
	res *meshtrust_api.MeshTrustResource
}

func MeshTrust() *MeshTrustBuilder {
	return &MeshTrustBuilder{
		res: &meshtrust_api.MeshTrustResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "trust-1",
			},
			Spec: &meshtrust_api.MeshTrust{
				TrustDomain: "default.east.mesh.local",
				CABundles: []meshtrust_api.CABundle{
					{
						Type: meshtrust_api.PemCABundleType,
						PEM: &meshtrust_api.PEM{
							Value: `-----BEGIN CERTIFICATE-----
MIIB9jCCAZ2gAwIBAgIUJ1gLZ/fvZhGq51qWrJzL6z2XWoQwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEChMHVGVzdCBDQTAeFw0yNTA3MzAxMjAwMDBaFw0zNTA3MjcxMjAw
MDBaMBIxEDAOBgNVBAoTB1Rlc3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AAQaXDFzOPslZ4e8n2KjsNkG+Wxi37L2zRWdMczDi7VqLrO03lczkB98/vzrdMKF
JgYxycULx10EYgMGQkgLrf1po2QwYjAdBgNVHQ4EFgQUUQBd5VjEO3N4XcgrxgMK
NU9xIQswHwYDVR0jBBgwFoAUUQBd5VjEO3N4XcgrxgMKNU9xIQswDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiA3dhhIQCzNkeGSjj6jK+jGE8fEKVmp
c9Vh+kJkmPUJZQIgQBr2GkV8uSfq/5ZKHD6jz6MJvKsg06dMBdvZBIA2ujg=
-----END CERTIFICATE-----`,
						},
					},
				},
			},
		},
	}
}

func (mi *MeshTrustBuilder) Build() *meshtrust_api.MeshTrustResource {
	return mi.res
}

func (mi *MeshTrustBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), mi.Build(), store.CreateBy(model.MetaToResourceKey(mi.res.GetMeta())))
}

func (mi *MeshTrustBuilder) WithName(name string) *MeshTrustBuilder {
	mi.res.Meta.(*test_model.ResourceMeta).Name = name
	return mi
}

func (mi *MeshTrustBuilder) WithTrustDomain(td string) *MeshTrustBuilder {
	mi.res.Spec.TrustDomain = td
	return mi
}

func (mi *MeshTrustBuilder) WithMesh(name string) *MeshTrustBuilder {
	mi.res.Meta.(*test_model.ResourceMeta).Mesh = name
	return mi
}
