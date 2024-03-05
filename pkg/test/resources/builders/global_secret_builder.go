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

type GlobalSecretBuilder struct {
	res *system.GlobalSecretResource
}

func GlobalSecret() *GlobalSecretBuilder {
	return &GlobalSecretBuilder{
		res: &system.GlobalSecretResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
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

func (s *GlobalSecretBuilder) Build() *system.GlobalSecretResource {
	return s.res
}

func (s *GlobalSecretBuilder) Create(st store.ResourceStore) error {
	return st.Create(context.Background(), s.Build(), store.CreateBy(s.Key()))
}

func (s *GlobalSecretBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(s.res.GetMeta())
}

func (s *GlobalSecretBuilder) WithName(name string) *GlobalSecretBuilder {
	s.res.Meta.(*test_model.ResourceMeta).Name = name
	return s
}

func (s *GlobalSecretBuilder) WithStringValue(val string) *GlobalSecretBuilder {
	s.res.Spec.Data.Value = []byte(val)
	return s
}
