package v1alpha1

import (
	"encoding/json"
	"fmt"
)

// Protobuf wrote an enum as the name of the value rather than its number, and left the
// zero value out of the document entirely. The numbers are kept because the zero value
// carries meaning: an inbound with no state recorded is Ready.
func marshalEnum(value int32, names map[int32]string) ([]byte, error) {
	name, ok := names[value]
	if !ok {
		return nil, fmt.Errorf("unknown enum value %d", value)
	}
	return json.Marshal(name)
}

// unmarshalEnum accepts the name and the number, which is what jsonpb read.
func unmarshalEnum(data []byte, values map[string]int32) (int32, error) {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		value, ok := values[name]
		if !ok {
			return 0, fmt.Errorf("unknown enum value %q", name)
		}
		return value, nil
	}
	var number int32
	if err := json.Unmarshal(data, &number); err != nil {
		return 0, fmt.Errorf("enum must be a name or a number: %w", err)
	}
	return number, nil
}

// Dataplane_Networking_Inbound_State describes whether an inbound serves traffic.
// Ready serves it, NotReady does not, and Ignored is not created at all: it cannot be
// targeted by policies, though the proxy still receives a certificate with its identity.
type Dataplane_Networking_Inbound_State int32

const (
	Dataplane_Networking_Inbound_Ready    Dataplane_Networking_Inbound_State = 0
	Dataplane_Networking_Inbound_NotReady Dataplane_Networking_Inbound_State = 1
	Dataplane_Networking_Inbound_Ignored  Dataplane_Networking_Inbound_State = 2
)

var (
	Dataplane_Networking_Inbound_State_name = map[int32]string{
		0: "Ready",
		1: "NotReady",
		2: "Ignored",
	}
	Dataplane_Networking_Inbound_State_value = map[string]int32{
		"Ready":    0,
		"NotReady": 1,
		"Ignored":  2,
	}
)

func (s Dataplane_Networking_Inbound_State) String() string {
	return Dataplane_Networking_Inbound_State_name[int32(s)]
}

func (s Dataplane_Networking_Inbound_State) MarshalJSON() ([]byte, error) {
	return marshalEnum(int32(s), Dataplane_Networking_Inbound_State_name)
}

func (s *Dataplane_Networking_Inbound_State) UnmarshalJSON(data []byte) error {
	value, err := unmarshalEnum(data, Dataplane_Networking_Inbound_State_value)
	if err != nil {
		return err
	}
	*s = Dataplane_Networking_Inbound_State(value)
	return nil
}

// Dataplane_Networking_Listener_Type distinguishes a zone ingress listener from a zone
// egress one on a proxy that carries both.
type Dataplane_Networking_Listener_Type int32

const (
	Dataplane_Networking_Listener_Unspecified Dataplane_Networking_Listener_Type = 0
	Dataplane_Networking_Listener_ZoneIngress Dataplane_Networking_Listener_Type = 1
	Dataplane_Networking_Listener_ZoneEgress  Dataplane_Networking_Listener_Type = 2
)

var (
	Dataplane_Networking_Listener_Type_name = map[int32]string{
		0: "Unspecified",
		1: "ZoneIngress",
		2: "ZoneEgress",
	}
	Dataplane_Networking_Listener_Type_value = map[string]int32{
		"Unspecified": 0,
		"ZoneIngress": 1,
		"ZoneEgress":  2,
	}
)

func (t Dataplane_Networking_Listener_Type) String() string {
	return Dataplane_Networking_Listener_Type_name[int32(t)]
}

func (t Dataplane_Networking_Listener_Type) MarshalJSON() ([]byte, error) {
	return marshalEnum(int32(t), Dataplane_Networking_Listener_Type_name)
}

func (t *Dataplane_Networking_Listener_Type) UnmarshalJSON(data []byte) error {
	value, err := unmarshalEnum(data, Dataplane_Networking_Listener_Type_value)
	if err != nil {
		return err
	}
	*t = Dataplane_Networking_Listener_Type(value)
	return nil
}

// Dataplane_Networking_Listener_State describes whether a listener serves traffic.
type Dataplane_Networking_Listener_State int32

const (
	Dataplane_Networking_Listener_Ready    Dataplane_Networking_Listener_State = 0
	Dataplane_Networking_Listener_NotReady Dataplane_Networking_Listener_State = 1
)

var (
	Dataplane_Networking_Listener_State_name = map[int32]string{
		0: "Ready",
		1: "NotReady",
	}
	Dataplane_Networking_Listener_State_value = map[string]int32{
		"Ready":    0,
		"NotReady": 1,
	}
)

func (s Dataplane_Networking_Listener_State) String() string {
	return Dataplane_Networking_Listener_State_name[int32(s)]
}

func (s Dataplane_Networking_Listener_State) MarshalJSON() ([]byte, error) {
	return marshalEnum(int32(s), Dataplane_Networking_Listener_State_name)
}

func (s *Dataplane_Networking_Listener_State) UnmarshalJSON(data []byte) error {
	value, err := unmarshalEnum(data, Dataplane_Networking_Listener_State_value)
	if err != nil {
		return err
	}
	*s = Dataplane_Networking_Listener_State(value)
	return nil
}
