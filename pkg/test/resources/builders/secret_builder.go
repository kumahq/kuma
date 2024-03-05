package builders

import (
	"context"

	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type SecretBuilder struct {
	res *system.SecretResource
}

func Secret() *SecretBuilder {
	return &SecretBuilder{
		res: &system.SecretResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "sec-1",
			},
			Spec: &system_proto.Secret{
				Data: &wrapperspb.BytesValue{
					Value: []byte("XYZ"),
				},
			},
		},
	}
}

func (s *SecretBuilder) Build() *system.SecretResource {
	return s.res
}

func (s *SecretBuilder) Create(st store.ResourceStore) error {
	return st.Create(context.Background(), s.Build(), store.CreateBy(s.Key()))
}

func (s *SecretBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(s.res.GetMeta())
}

func (s *SecretBuilder) WithName(name string) *SecretBuilder {
	s.res.Meta.(*test_model.ResourceMeta).Name = name
	return s
}

func (s *SecretBuilder) WithMesh(mesh string) *SecretBuilder {
	s.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return s
}

func (s *SecretBuilder) WithStringValue(val string) *SecretBuilder {
	s.res.Spec.Data.Value = []byte(val)
	return s
}
