package vips

import (
	"fmt"
	"net"
	"sort"
)

type VirtualOutboundMeshView struct {
	byHostname map[HostnameEntry]*VirtualOutbound
}

func NewEmptyVirtualOutboundView() *VirtualOutboundMeshView {
	return &VirtualOutboundMeshView{
		byHostname: map[HostnameEntry]*VirtualOutbound{},
	}
}

func NewVirtualOutboundView(all map[HostnameEntry]VirtualOutbound) (*VirtualOutboundMeshView, error) {
	r := NewEmptyVirtualOutboundView()
	for k := range all {
		itm := all[k]
		r.byHostname[k] = &itm
		if len(itm.Outbounds) == 0 {
			return nil, fmt.Errorf("no outbound for hostname: %s", k)
		}
	}
	return r, nil
}

func (vo *VirtualOutboundMeshView) Get(entry HostnameEntry) *VirtualOutbound {
	return vo.byHostname[entry]
}

func (vo *VirtualOutboundMeshView) Add(entry HostnameEntry, outbound OutboundEntry) error {
	if vo.byHostname[entry] == nil {
		vo.byHostname[entry] = &VirtualOutbound{Outbounds: []OutboundEntry{outbound}}
		if entry.Type == Host && net.ParseIP(entry.Name) != nil {
			// Entry name can be IP in case of providing IP in ExternalService or building VIPs based on ClusterIP/PodIP on Kubernetes
			vo.byHostname[entry].Address = entry.Name
		}
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

func (vo *VirtualOutboundMeshView) HostnameEntries() []HostnameEntry {
	keys := make([]HostnameEntry, 0, len(vo.byHostname))
	for k := range vo.byHostname {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].Less(&keys[j])
	})
	return keys
}

// Update merges `new` and `vo` in a new `out` and returns a list of changes.
func (vo *VirtualOutboundMeshView) Update(new *VirtualOutboundMeshView) (changes []Change, out *VirtualOutboundMeshView) {
	changes = []Change{}
	out = NewEmptyVirtualOutboundView()
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

func (vo *VirtualOutboundMeshView) DeleteByOrigin(origin string) {
	for entry := range vo.byHostname {
		if len(vo.byHostname[entry].Outbounds) > 0 && vo.byHostname[entry].Outbounds[0].Origin == origin {
			delete(vo.byHostname, entry)
		}
	}
}

type ChangeType string

const (
	Add    = ChangeType("Add")
	Remove = ChangeType("Remove")
	Modify = ChangeType("Modify")
)

type Change struct {
	Type  ChangeType
	Entry HostnameEntry
}

func (c Change) String() string {
	return fmt.Sprintf("%s %s", c.Type, c.Entry.String())
}
