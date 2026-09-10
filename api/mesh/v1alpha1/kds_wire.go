package v1alpha1

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/v3/api/mesh/v1alpha1/kdswire"
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
