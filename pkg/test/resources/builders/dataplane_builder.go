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

var (
	FirstInboundPort        = uint32(80)
	FirstInboundServicePort = uint32(8080)
	FirstOutboundPort       = uint32(10001)
)

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
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
				},
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

func (d *DataplaneBuilder) WithHttpServices(services ...string) *DataplaneBuilder {
	d.WithoutInbounds()
	for _, service := range services {
		d.AddInboundHttpOfService(service)
	}
	return d
}

func (d *DataplaneBuilder) WithoutInbounds() *DataplaneBuilder {
	d.res.Spec.Networking.Inbound = nil
	return d
}

func (d *DataplaneBuilder) WithInboundOfTags(tagsKV ...string) *DataplaneBuilder {
	return d.WithInboundOfTagsMap(TagsKVToMap(tagsKV))
}

func (d *DataplaneBuilder) WithInboundOfTagsMap(tags map[string]string) *DataplaneBuilder {
	return d.WithoutInbounds().AddInboundOfTagsMap(tags)
}

func (d *DataplaneBuilder) AddInboundOfService(service string) *DataplaneBuilder {
	return d.AddInboundOfTags(mesh_proto.ServiceTag, service)
}

func (d *DataplaneBuilder) AddInboundHttpOfService(service string) *DataplaneBuilder {
	return d.AddInboundOfTags(mesh_proto.ServiceTag, service, mesh_proto.ProtocolTag, "http")
}

func (d *DataplaneBuilder) AddInboundOfTags(tags ...string) *DataplaneBuilder {
	return d.AddInboundOfTagsMap(TagsKVToMap(tags))
}

func (d *DataplaneBuilder) AddInboundOfTagsMap(tags map[string]string) *DataplaneBuilder {
	return d.AddInbound(
		Inbound().
			WithPort(FirstInboundPort + uint32(len(d.res.Spec.Networking.Inbound))).
			WithServicePort(FirstInboundServicePort + uint32(len(d.res.Spec.Networking.Inbound))).
			WithTags(tags),
	)
}

func (d *DataplaneBuilder) AddInbound(inbound *InboundBuilder) *DataplaneBuilder {
	d.res.Spec.Networking.Inbound = append(d.res.Spec.Networking.Inbound, inbound.Build())
	return d
}

func (d *DataplaneBuilder) AddOutbound(outbound *OutboundBuilder) *DataplaneBuilder {
	d.res.Spec.Networking.Outbound = append(d.res.Spec.Networking.Outbound, outbound.Build())
	return d
}

func (d *DataplaneBuilder) AddOutbounds(outbounds []*OutboundBuilder) *DataplaneBuilder {
	for _, outbound := range outbounds {
		d.res.Spec.Networking.Outbound = append(d.res.Spec.Networking.Outbound, outbound.Build())
	}
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

func (d *DataplaneBuilder) WithTransparentProxying(redirectPortOutbound, redirectPortInbound uint32, ipFamilyMode string) *DataplaneBuilder {
	d.res.Spec.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
		RedirectPortInbound:  redirectPortInbound,
		RedirectPortOutbound: redirectPortOutbound,
		IpFamilyMode:         ipFamilyModeEnumValue(ipFamilyMode),
	}
	return d
}

func ipFamilyModeEnumValue(mode string) mesh_proto.Dataplane_Networking_TransparentProxying_IpFamilyMode {
	switch mode {
	case "ipv4":
		return mesh_proto.Dataplane_Networking_TransparentProxying_IPv4
	case "dualstack":
		fallthrough
	case "ipv6":
		fallthrough
	default:
		return mesh_proto.Dataplane_Networking_TransparentProxying_DualStack
	}
}

func TagsKVToMap(tagsKV []string) map[string]string {
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

func (d *DataplaneBuilder) WithBuiltInGateway(name string) *DataplaneBuilder {
	d.res.Spec.Networking.Gateway = &mesh_proto.Dataplane_Networking_Gateway{
		Tags: map[string]string{
			mesh_proto.ServiceTag: name,
		},
		Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
	}
	return d
}

func (d *DataplaneBuilder) AddBuiltInGatewayTags(tags map[string]string) *DataplaneBuilder {
	for k, v := range tags {
		d.res.Spec.Networking.Gateway.Tags[k] = v
	}
	return d
}

func (d *DataplaneBuilder) WithAdminPort(i int) *DataplaneBuilder {
	d.res.Spec.Networking.Admin = &mesh_proto.EnvoyAdmin{
		Port: uint32(i),
	}
	return d
}

type InboundBuilder struct {
	res *mesh_proto.Dataplane_Networking_Inbound
}

func Inbound() *InboundBuilder {
	return &InboundBuilder{
		res: &mesh_proto.Dataplane_Networking_Inbound{
			Tags: map[string]string{},
		},
	}
}

func (b *InboundBuilder) WithAddress(addr string) *InboundBuilder {
	b.res.Address = addr
	return b
}

func (b *InboundBuilder) WithPort(port uint32) *InboundBuilder {
	b.res.Port = port
	return b
}

func (b *InboundBuilder) WithName(name string) *InboundBuilder {
	b.res.Name = name
	return b
}

func (b *InboundBuilder) WithServicePort(port uint32) *InboundBuilder {
	b.res.ServicePort = port
	return b
}

func (b *InboundBuilder) WithTags(tags map[string]string) *InboundBuilder {
	for k, v := range tags {
		b.res.Tags[k] = v
	}
	return b
}

func (b *InboundBuilder) WithService(name string) *InboundBuilder {
	b.WithTags(map[string]string{mesh_proto.ServiceTag: name})
	return b
}

func (b *InboundBuilder) Build() *mesh_proto.Dataplane_Networking_Inbound {
	return b.res
}

type OutboundBuilder struct {
	res *mesh_proto.Dataplane_Networking_Outbound
}

func Outbound() *OutboundBuilder {
	return &OutboundBuilder{
		res: &mesh_proto.Dataplane_Networking_Outbound{
			Tags: map[string]string{},
		},
	}
}

func (b *OutboundBuilder) WithAddress(addr string) *OutboundBuilder {
	b.res.Address = addr
	return b
}

func (b *OutboundBuilder) WithPort(port uint32) *OutboundBuilder {
	b.res.Port = port
	return b
}

func (b *OutboundBuilder) WithTags(tags map[string]string) *OutboundBuilder {
	for k, v := range tags {
		b.res.Tags[k] = v
	}
	return b
}

func (b *OutboundBuilder) WithService(name string) *OutboundBuilder {
	b.WithTags(map[string]string{mesh_proto.ServiceTag: name})
	return b
}

func (b *OutboundBuilder) WithMeshService(name string, port uint32) *OutboundBuilder {
	b.res.Tags = nil
	b.res.BackendRef = &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
		Kind: "MeshService",
		Name: name,
		Port: port,
	}
	return b
}

func (b *OutboundBuilder) WithMeshExternalService(name string, port uint32) *OutboundBuilder {
	b.res.Tags = nil
	b.res.BackendRef = &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
		Kind: "MeshExternalService",
		Name: name,
		Port: port,
	}
	return b
}

func (b *OutboundBuilder) WithMeshMultiZoneService(name string, port uint32) *OutboundBuilder {
	b.res.Tags = nil
	b.res.BackendRef = &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
		Kind: "MeshMultiZoneService",
		Name: name,
		Port: port,
	}
	return b
}

func (b *OutboundBuilder) Build() *mesh_proto.Dataplane_Networking_Outbound {
	return b.res
}
