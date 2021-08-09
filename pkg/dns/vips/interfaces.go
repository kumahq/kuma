package vips

import (
	"fmt"
	"sort"
)

type EntryType int

const (
	Service EntryType = iota
	Host
)

func (t EntryType) String() string {
	switch t {
	case Service:
		return "service"
	case Host:
		return "host"
	default:
		return "undefined"
	}
}

type Entry struct {
	Type EntryType `json:"type"`
	Name string    `json:"name"`
}

func (e Entry) String() string {
	return fmt.Sprintf("%s:%s", e.Type, e.Name)
}

func (e Entry) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%d:%s", e.Type, e.Name)), nil
}

func (e *Entry) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "%v:%s", &e.Type, &e.Name)
	return err
}

func NewHostEntry(host string) Entry {
	return Entry{Host, host}
}

func NewServiceEntry(name string) Entry {
	return Entry{Service, name}
}

type EntrySet map[Entry]bool

func (s EntrySet) ToArray() (entries []Entry) {
	for entry := range s {
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		// Sort by entry type, then name.
		return entries[i].Type < entries[j].Type ||
			entries[i].Name < entries[j].Name
	})
	return
}

type List map[Entry]string

func (vips List) Append(other List) {
	for k, v := range other {
		vips[k] = v
	}
}

func (vips List) FQDNsByIPs() map[string]Entry {
	ipToDomain := map[string]Entry{}
	for domain, ip := range vips {
		ipToDomain[ip] = domain
	}
	return ipToDomain
}
