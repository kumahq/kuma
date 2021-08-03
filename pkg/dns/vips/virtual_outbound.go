package vips

import (
	"fmt"
	"sort"
	"strings"
)

// VirtualOutbound the description of a hostname --> address and later port/tagSet
type VirtualOutbound struct {
	// This is not default in the legacy case (hostnames won't be complete)
	Address   string         `json:"address,omitempty"`
	Outbounds []MeshOutbound `json:"outbounds,omitempty"`
}

func (vo *VirtualOutbound) Equal(other *VirtualOutbound) bool {
	if vo.Address != other.Address || len(vo.Outbounds) != len(other.Outbounds) {
		return false
	}
	for i := range vo.Outbounds {
		if vo.Outbounds[i].String() != other.Outbounds[i].String() {
			return false
		}
	}
	return true
}

type MeshOutbound struct {
	Port   uint32
	TagSet map[string]string
	// A string to identify where this outbound was defined (usually the name of the outbound policy)
	Origin string
}

func (mo *MeshOutbound) Less(o *MeshOutbound) bool {
	return mo.Port < o.Port
}

func (mo *MeshOutbound) String() string {
	var tags []string
	for k, v := range mo.TagSet {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(tags)
	return fmt.Sprintf("%s=%d{%s}", mo.Origin, mo.Port, strings.Join(tags, ","))
}

type VirtualOutboundView struct {
	byHostname map[Entry]*VirtualOutbound
}

func NewVirtualOutboundView(all map[Entry]VirtualOutbound) *VirtualOutboundView {
	r := VirtualOutboundView{
		byHostname: map[Entry]*VirtualOutbound{},
	}
	for k := range all {
		itm := all[k]
		r.byHostname[k] = &itm
	}
	return &r
}

func (vo *VirtualOutboundView) Get(entry Entry) *VirtualOutbound {
	return vo.byHostname[entry]
}

func (vo *VirtualOutboundView) Add(entry Entry, outbound MeshOutbound) error {
	if vo.byHostname[entry] == nil {
		vo.byHostname[entry] = &VirtualOutbound{Outbounds: []MeshOutbound{outbound}}
		return nil
	}
	for _, existingOutbound := range vo.byHostname[entry].Outbounds {
		if existingOutbound.Port == outbound.Port {
			if existingOutbound.String() == outbound.String() {
				return nil
			}
			return fmt.Errorf("can't add %s:%d from %s because it's already used by entity defined in:'%s'", entry.Name, outbound.Port, outbound.Origin, existingOutbound.Origin)
		}
	}
	vo.byHostname[entry].Outbounds = append(vo.byHostname[entry].Outbounds, outbound)
	sort.SliceStable(vo.byHostname[entry].Outbounds, func(i, j int) bool {
		return vo.byHostname[entry].Outbounds[i].Less(&vo.byHostname[entry].Outbounds[j])
	})
	return nil
}

func (vo *VirtualOutboundView) Keys() []Entry {
	keys := make([]Entry, 0, len(vo.byHostname))
	for k := range vo.byHostname {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].Less(&keys[j])
	})
	return keys
}

// Update merges `new` and `vo` in a new `out` and return a list of changes.
func (vo *VirtualOutboundView) Update(new *VirtualOutboundView) (changes []Change, out *VirtualOutboundView) {
	changes = []Change{}
	out = NewVirtualOutboundView(map[Entry]VirtualOutbound{})
	// Let's find the removed ones (in old but not in new)
	for entry := range vo.byHostname {
		if _, ok := new.byHostname[entry]; !ok {
			changes = append(changes, Change{Type: Remove, Entry: entry})
		}
	}
	for entry, vob := range new.byHostname {
		oldVob, ok := vo.byHostname[entry]
		if ok {
			if !oldVob.Equal(vob) {
				changes = append(changes, Change{Type: Modify, Entry: entry})
			}
		} else {
			changes = append(changes, Change{Type: Add, Entry: entry})
		}
		out.byHostname[entry] = &VirtualOutbound{Address: vob.Address, Outbounds: vob.Outbounds}
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Entry == changes[j].Entry {
			return changes[i].Type == Add
		}
		return changes[i].Entry.String() < changes[j].Entry.String()
	})
	return
}

type ChangeType string

const (
	Add    = ChangeType("Add")
	Remove = ChangeType("Remove")
	Modify = ChangeType("Modify")
)

type Change struct {
	Type  ChangeType
	Entry Entry
}

func (c Change) String() string {
	return fmt.Sprintf("%s %s", c.Type, c.Entry.String())
}
