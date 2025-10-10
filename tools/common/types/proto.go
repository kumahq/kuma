package types

import (
	"reflect"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
)

var ProtoTypeToType = map[string]reflect.Type{
	"Mesh":        reflect.TypeOf(v1alpha1.Mesh{}),
	"Dataplane":   reflect.TypeOf(v1alpha1.Dataplane{}),
	"MeshGateway": reflect.TypeOf(v1alpha1.MeshGateway{}),
	"ZoneIngress": reflect.TypeOf(v1alpha1.ZoneIngress{}),
	"ZoneEgress":  reflect.TypeOf(v1alpha1.ZoneEgress{}),
}
