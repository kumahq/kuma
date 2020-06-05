package resolver

type DNSResolver interface {
	Start(<-chan struct{}) error
	AddDomain(domain string) error
	RemoveDomain(domain string) error
	AddServiceToDomain(service string, domain string) (string, error)
	RemoveServiceFromDomain(service string, domain string) error
	ForwardLookup(name string) (string, error)
	ReverseLookup(ip string) (string, error)
}
