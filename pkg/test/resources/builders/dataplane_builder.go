package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var FirstInboundPort = uint32(80)
var FirstInboundServicePort = uint32(8080)
var FirstOutboundPort = uint32(10001)

type DataplaneBuilder struct {
	res *core_mesh.DataplaneResource
}

func Dataplane() *DataplaneBuilder {
	return &DataplaneBuilder{
		res: &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "dp-1",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{},
			},
		},
	}
}

func (d *DataplaneBuilder) Build() *core_mesh.DataplaneResource {
	if err := d.res.Validate(); err != nil {
		panic(err)
	}
	return d.res
}

func (d *DataplaneBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), d.Build(), store.CreateBy(d.Key()))
}

func (d *DataplaneBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(d.res.GetMeta())
}

func (d *DataplaneBuilder) With(fn func(*core_mesh.DataplaneResource)) *DataplaneBuilder {
	fn(d.res)
	return d
}

func (d *DataplaneBuilder) WithName(name string) *DataplaneBuilder {
	d.res.Meta.(*test_model.ResourceMeta).Name = name
	return d
}

func (d *DataplaneBuilder) WithMesh(mesh string) *DataplaneBuilder {
	d.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return d
}

func (d *DataplaneBuilder) WithVersion(version string) *DataplaneBuilder {
	d.res.Meta.(*test_model.ResourceMeta).Version = version
	return d
}

func (d *DataplaneBuilder) WithAddress(address string) *DataplaneBuilder {
	d.res.Spec.Networking.Address = address
	return d
}

func (d *DataplaneBuilder) WithServices(services ...string) *DataplaneBuilder {
	d.WithoutInbounds()
	for _, service := range services {
		d.AddInboundOfService(service)
	}
	return d
}

func (d *DataplaneBuilder) WithoutInbounds() *DataplaneBuilder {
	d.res.Spec.Networking.Inbound = nil
	return d
}

func (d *DataplaneBuilder) WithInboundOfTags(tagsKV ...string) *DataplaneBuilder {
	return d.WithInboundOfTagsMap(tagsKVToMap(tagsKV))
}

func (d *DataplaneBuilder) WithInboundOfTagsMap(tags map[string]string) *DataplaneBuilder {
	return d.WithoutInbounds().AddInboundOfTagsMap(tags)
}

func (d *DataplaneBuilder) AddInboundOfService(service string) *DataplaneBuilder {
	return d.AddInboundOfTags(mesh_proto.ServiceTag, service)
}

func (d *DataplaneBuilder) AddInboundOfTags(tags ...string) *DataplaneBuilder {
	return d.AddInboundOfTagsMap(tagsKVToMap(tags))
}

func (d *DataplaneBuilder) AddInboundOfTagsMap(tags map[string]string) *DataplaneBuilder {
	d.res.Spec.Networking.Inbound = append(d.res.Spec.Networking.Inbound, &mesh_proto.Dataplane_Networking_Inbound{
		Port:        FirstInboundPort + uint32(len(d.res.Spec.Networking.Inbound)),
		ServicePort: FirstInboundServicePort + uint32(len(d.res.Spec.Networking.Inbound)),
		Tags:        tags,
	})
	return d
}

func (d *DataplaneBuilder) AddOutboundToService(service string) *DataplaneBuilder {
	d.res.Spec.Networking.Outbound = append(d.res.Spec.Networking.Outbound, &mesh_proto.Dataplane_Networking_Outbound{
		Port: FirstOutboundPort + uint32(len(d.res.Spec.Networking.Outbound)),
		Tags: map[string]string{
			mesh_proto.ServiceTag: service,
		},
	})
	return d
}

func (d *DataplaneBuilder) AddOutboundsToServices(services ...string) *DataplaneBuilder {
	for _, service := range services {
		d.AddOutboundToService(service)
	}
	return d
}

func tagsKVToMap(tagsKV []string) map[string]string {
	if len(tagsKV)%2 == 1 {
		panic("tagsKV has to have even number of arguments")
	}
	tags := map[string]string{}
	for i := 0; i < len(tagsKV); i += 2 {
		tags[tagsKV[i]] = tagsKV[i+1]
	}
	return tags
}

func (d *DataplaneBuilder) WithPrometheusMetrics(config *mesh_proto.PrometheusMetricsBackendConfig) *DataplaneBuilder {
	d.res.Spec.Metrics = &mesh_proto.MetricsBackend{
		Name: "prometheus-1",
		Type: mesh_proto.MetricsPrometheusType,
		Conf: proto.MustToStruct(config),
	}
	return d
}
