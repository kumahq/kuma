package types

import (
	"reflect"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
)

var ProtoTypeToType = map[string]reflect.Type{
	"Mesh":        reflect.TypeFor[mesh_proto.Mesh](),
	"Secret":      reflect.TypeFor[system_proto.Secret](),
	"Dataplane":   reflect.TypeFor[mesh_proto.Dataplane](),
	"MeshGateway": reflect.TypeFor[mesh_proto.MeshGateway](),
	"ZoneIngress": reflect.TypeFor[mesh_proto.ZoneIngress](),
	"ZoneEgress":  reflect.TypeFor[mesh_proto.ZoneEgress](),
}
