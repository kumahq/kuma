package v1alpha1

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/v3/api/mesh/v1alpha1/kdswire"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

// Mesh was a protobuf message until Kuma 3.0 and is still that on the KDS wire, so
// that a zone control plane of any released version can read what a global control
// plane of this one sends. See kdswire/kds_wire.proto.
func (m *Mesh) ToKDSWire() proto.Message {
	return &kdswire.Mesh{SkipCreatingInitialPolicies: m.GetSkipCreatingInitialPolicies()}
}

func (m *Mesh) FromKDSWire(msg proto.Message) error {
	wire, ok := msg.(*kdswire.Mesh)
	if !ok {
		return fmt.Errorf("invalid type %T for the Mesh wire form", msg)
	}
	m.SkipCreatingInitialPolicies = wire.GetSkipCreatingInitialPolicies()
	return nil
}

// Dataplane and DataplaneInsight were protobuf messages until Kuma 3.0 and are still
// those on the KDS wire. Their trees are large, so rather than restate every field the
// conversion goes through the JSON form, which the compatibility tests hold byte
// identical to what protobuf wrote. A conversion that cannot happen leaves the wire
// message empty rather than sending a half filled one.
func (d *Dataplane) ToKDSWire() proto.Message {
	wire := &kdswire.Dataplane{}
	if err := toWire(d, wire); err != nil {
		return &kdswire.Dataplane{}
	}
	return wire
}

func (d *Dataplane) FromKDSWire(msg proto.Message) error {
	wire, ok := msg.(*kdswire.Dataplane)
	if !ok {
		return fmt.Errorf("invalid type %T for the Dataplane wire form", msg)
	}
	return fromWire(wire, d)
}

func (i *DataplaneInsight) ToKDSWire() proto.Message {
	wire := &kdswire.DataplaneInsight{}
	if err := toWire(i, wire); err != nil {
		return &kdswire.DataplaneInsight{}
	}
	return wire
}

func (i *DataplaneInsight) FromKDSWire(msg proto.Message) error {
	wire, ok := msg.(*kdswire.DataplaneInsight)
	if !ok {
		return fmt.Errorf("invalid type %T for the DataplaneInsight wire form", msg)
	}
	return fromWire(wire, i)
}

func toWire(spec any, wire proto.Message) error {
	encoded, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return util_proto.FromJSON(encoded, wire)
}

func fromWire(wire proto.Message, spec any) error {
	encoded, err := util_proto.ToJSON(wire)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, spec)
}
