package builders

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	meshtrust_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

type MeshTrustBuilder struct {
	res       *meshtrust_api.MeshTrustResource
	namespace string
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
			Status: &meshtrust_api.MeshTrustStatus{},
		},
	}
}

func (mtr *MeshTrustBuilder) Build() *meshtrust_api.MeshTrustResource {
	return mtr.res
}

func (mtr *MeshTrustBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), mtr.Build(), store.CreateBy(model.MetaToResourceKey(mtr.res.GetMeta())))
}

func (mtr *MeshTrustBuilder) WithName(name string) *MeshTrustBuilder {
	mtr.res.Meta.(*test_model.ResourceMeta).Name = name
	return mtr
}

func (mtr *MeshTrustBuilder) WithTrustDomain(td string) *MeshTrustBuilder {
	mtr.res.Spec.TrustDomain = td
	return mtr
}

func (mtr *MeshTrustBuilder) WithNamespace(ns string) *MeshTrustBuilder {
	mtr.namespace = ns
	return mtr
}

func (mtr *MeshTrustBuilder) WithMesh(name string) *MeshTrustBuilder {
	mtr.res.Meta.(*test_model.ResourceMeta).Mesh = name
	return mtr
}

func (mtr *MeshTrustBuilder) WithCA(ca string) *MeshTrustBuilder {
	mtr.res.Spec.CABundles = append([]meshtrust_api.CABundle{}, meshtrust_api.CABundle{
		Type: meshtrust_api.PemCABundleType,
		PEM: &meshtrust_api.PEM{
			Value: ca,
		},
	})
	return mtr
}

func (mtr *MeshTrustBuilder) WithLabels(labels map[string]string) *MeshTrustBuilder {
	if mtr.res.Meta.(*test_model.ResourceMeta).Labels == nil {
		mtr.res.Meta.(*test_model.ResourceMeta).Labels = map[string]string{}
	}
	for key, val := range labels {
		mtr.res.Meta.(*test_model.ResourceMeta).Labels[key] = val
	}
	return mtr
}

func (mtr *MeshTrustBuilder) KubeYaml() string {
	trust := mtr.Build()
	labels := trust.Meta.GetLabels()
	labels[mesh_proto.MeshTag] = trust.Meta.GetMesh()
	kubeTrust := meshtrust_k8s.MeshTrust{
		TypeMeta: v1.TypeMeta{
			Kind:       string(meshtrust_api.MeshTrustType),
			APIVersion: mesh_k8s.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: mtr.namespace,
			Name:      trust.Meta.GetName(),
			Labels:    labels,
		},
	}
	kubeTrust.SetSpec(trust.Spec)
	res, err := yaml.Marshal(kubeTrust)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func (mtr *MeshTrustBuilder) UniYaml() string {
	trust := mtr.Build()
	res, err := yaml.Marshal(rest.From.Resource(trust))
	if err != nil {
		panic(err)
	}
	return string(res)
}
