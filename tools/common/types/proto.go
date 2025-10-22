package types

import (
	"reflect"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
)

var ProtoTypeToType = map[string]reflect.Type{
	"Mesh":        reflect.TypeOf(mesh_proto.Mesh{}),
	"Secret":      reflect.TypeOf(system_proto.Secret{}),
	"Dataplane":   reflect.TypeOf(mesh_proto.Dataplane{}),
	"MeshGateway": reflect.TypeOf(mesh_proto.MeshGateway{}),
	"ZoneIngress": reflect.TypeOf(mesh_proto.ZoneIngress{}),
	"ZoneEgress":  reflect.TypeOf(mesh_proto.ZoneEgress{}),
}
