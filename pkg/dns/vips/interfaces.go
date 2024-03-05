package vips

import (
	"fmt"

	"github.com/pkg/errors"
)

type EntryType int

const (
	Service EntryType = iota
	Host
	FullyQualifiedDomain
)

func (t EntryType) String() string {
	switch t {
	case Service:
		return "service"
	case Host:
		return "host"
	case FullyQualifiedDomain:
		return "fqdn"
	default:
		return "undefined"
	}
}

// HostnameEntry is the definition of a DNS entry. The type indicates where the entry comes from
// (.e.g: Service and Host are auto-generated, FullyQualifiedDomain comes from `virtual-outbound` policies...)
type HostnameEntry struct {
	Type EntryType `json:"type"`
	Name string    `json:"name"`
}

func (e HostnameEntry) String() string {
	return fmt.Sprintf("%s:%s", e.Type, e.Name)
}

func (e HostnameEntry) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d:%s", e.Type, e.Name)), nil
}

func (e *HostnameEntry) UnmarshalText(text []byte) error {
	if _, err := fmt.Sscanf(string(text), "%v:%s", &e.Type, &e.Name); err != nil {
		return errors.Wrapf(err, "could not parse hostname entry: %+q", text)
	}

	return nil
}

func (e *HostnameEntry) Less(o *HostnameEntry) bool {
	if e.Type == o.Type {
		return e.Name < o.Name
	}
	return e.Type < o.Type
}

func NewHostEntry(host string) HostnameEntry {
	return HostnameEntry{Host, host}
}

func NewServiceEntry(name string) HostnameEntry {
	return HostnameEntry{Service, name}
}

func NewFqdnEntry(name string) HostnameEntry {
	return HostnameEntry{FullyQualifiedDomain, name}
}
