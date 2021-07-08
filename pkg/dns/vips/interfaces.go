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

type Entry struct {
	Type EntryType `json:"type"`
	Name string    `json:"name"`
}

func (e Entry) String() string {
	return fmt.Sprintf("%v:%s", e.Type, e.Name)
}

func (e Entry) MarshalText() (text []byte, err error) {
	return []byte(e.String()), nil
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
		return entries[i].String() < entries[j].String()
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
