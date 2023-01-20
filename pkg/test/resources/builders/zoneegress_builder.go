package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type ZoneEgressBuilder struct {
	res *core_mesh.ZoneEgressResource
}

func ZoneEgress() *ZoneEgressBuilder {
	return &ZoneEgressBuilder{
		res: &core_mesh.ZoneEgressResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "zoneegress-1",
			},
			Spec: &mesh_proto.ZoneEgress{
				Networking: &mesh_proto.ZoneEgress_Networking{
					Address: "127.0.0.1",
				},
			},
		},
	}
}

func (b *ZoneEgressBuilder) Build() *core_mesh.ZoneEgressResource {
	if err := b.res.Validate(); err != nil {
		panic(err)
	}
	return b.res
}

func (b *ZoneEgressBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), b.Build(), store.CreateBy(b.Key()))
}

func (b *ZoneEgressBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(b.res.GetMeta())
}

func (b *ZoneEgressBuilder) With(fn func(*core_mesh.ZoneEgressResource)) *ZoneEgressBuilder {
	fn(b.res)
	return b
}

func (b *ZoneEgressBuilder) WithName(name string) *ZoneEgressBuilder {
	b.res.Meta.(*test_model.ResourceMeta).Name = name
	return b
}

func (b *ZoneEgressBuilder) WithVersion(version string) *ZoneEgressBuilder {
	b.res.Meta.(*test_model.ResourceMeta).Version = version
	return b
}

func (b *ZoneEgressBuilder) WithAddress(address string) *ZoneEgressBuilder {
	b.res.Spec.Networking.Address = address
	return b
}

func (b *ZoneEgressBuilder) WithZone(zone string) *ZoneEgressBuilder {
	b.res.Spec.Zone = zone
	return b
}

func (b *ZoneEgressBuilder) WithAdminPort(i int) *ZoneEgressBuilder {
	b.res.Spec.Networking.Admin = &mesh_proto.EnvoyAdmin{
		Port: uint32(i),
	}
	return b
}

func (b *ZoneEgressBuilder) WithPort(port uint32) *ZoneEgressBuilder {
	b.res.Spec.Networking.Port = port
	return b
}
