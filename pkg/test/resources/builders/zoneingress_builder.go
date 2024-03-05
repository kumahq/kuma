package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type ZoneIngressBuilder struct {
	res *core_mesh.ZoneIngressResource
}

func ZoneIngress() *ZoneIngressBuilder {
	return &ZoneIngressBuilder{
		res: &core_mesh.ZoneIngressResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "zoneingress-1",
			},
			Spec: &mesh_proto.ZoneIngress{
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address: "127.0.0.1",
					Port:    10000,
				},
			},
		},
	}
}

func (b *ZoneIngressBuilder) Build() *core_mesh.ZoneIngressResource {
	if err := b.res.Validate(); err != nil {
		panic(err)
	}
	return b.res
}

func (b *ZoneIngressBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), b.Build(), store.CreateBy(b.Key()))
}

func (b *ZoneIngressBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(b.res.GetMeta())
}

func (b *ZoneIngressBuilder) With(fn func(*core_mesh.ZoneIngressResource)) *ZoneIngressBuilder {
	fn(b.res)
	return b
}

func (b *ZoneIngressBuilder) WithName(name string) *ZoneIngressBuilder {
	b.res.Meta.(*test_model.ResourceMeta).Name = name
	return b
}

func (b *ZoneIngressBuilder) WithVersion(version string) *ZoneIngressBuilder {
	b.res.Meta.(*test_model.ResourceMeta).Version = version
	return b
}

func (b *ZoneIngressBuilder) WithAddress(address string) *ZoneIngressBuilder {
	b.res.Spec.Networking.Address = address
	return b
}

func (b *ZoneIngressBuilder) WithZone(zone string) *ZoneIngressBuilder {
	b.res.Spec.Zone = zone
	return b
}

func (b *ZoneIngressBuilder) WithAdminPort(i int) *ZoneIngressBuilder {
	b.res.Spec.Networking.Admin = &mesh_proto.EnvoyAdmin{
		Port: uint32(i),
	}
	return b
}

func (b *ZoneIngressBuilder) WithAdvertisedAddress(addr string) *ZoneIngressBuilder {
	b.res.Spec.Networking.AdvertisedAddress = addr
	return b
}

func (b *ZoneIngressBuilder) WithPort(port uint32) *ZoneIngressBuilder {
	b.res.Spec.Networking.Port = port
	return b
}

func (b *ZoneIngressBuilder) WithAdvertisedPort(port uint32) *ZoneIngressBuilder {
	b.res.Spec.Networking.AdvertisedPort = port
	return b
}

func (b *ZoneIngressBuilder) AddSimpleAvailableService(svc string) *ZoneIngressBuilder {
	return b.AddAvailableService(&mesh_proto.ZoneIngress_AvailableService{
		Tags: map[string]string{
			mesh_proto.ServiceTag: svc,
		},
		Instances:       1,
		Mesh:            "default",
		ExternalService: false,
	})
}

func (b *ZoneIngressBuilder) AddAvailableService(svc *mesh_proto.ZoneIngress_AvailableService) *ZoneIngressBuilder {
	b.res.Spec.AvailableServices = append(b.res.Spec.AvailableServices, svc)
	return b
}
