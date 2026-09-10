package v1alpha1

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kumahq/kuma/v3/api/system/v1alpha1/kdswire"
)

// Secret and Config were protobuf messages until Kuma 3.0 and are still that on
// the KDS wire, so that a zone control plane of any released version can read
// what a global control plane of this one sends. See kdswire/kds_wire.proto.
func (s *Secret) ToKDSWire() proto.Message {
	wire := &kdswire.Secret{}
	if s.Data != nil {
		wire.Data = wrapperspb.Bytes(s.Data.Value)
	}
	return wire
}

func (s *Secret) FromKDSWire(msg proto.Message) error {
	wire, ok := msg.(*kdswire.Secret)
	if !ok {
		return fmt.Errorf("invalid type %T for the Secret wire form", msg)
	}
	if wire.GetData() == nil {
		s.Data = nil
		return nil
	}
	s.Data = Bytes(wire.GetData().GetValue())
	return nil
}

func (c *Config) ToKDSWire() proto.Message {
	return &kdswire.Config{Config: c.Config}
}

func (c *Config) FromKDSWire(msg proto.Message) error {
	wire, ok := msg.(*kdswire.Config)
	if !ok {
		return fmt.Errorf("invalid type %T for the Config wire form", msg)
	}
	c.Config = wire.GetConfig()
	return nil
}
