package types

import (
	"reflect"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
)

var ProtoTypeToType = map[string]reflect.Type{
	"Mesh":             reflect.TypeFor[mesh_proto.Mesh](),
	"Secret":           reflect.TypeFor[system_proto.Secret](),
	"Dataplane":        reflect.TypeFor[mesh_proto.Dataplane](),
	"DataplaneInsight": reflect.TypeFor[mesh_proto.DataplaneInsight](),
	"Zone":             reflect.TypeFor[system_proto.Zone](),
	"ZoneInsight":      reflect.TypeFor[system_proto.ZoneInsight](),
}
