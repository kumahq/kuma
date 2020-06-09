package resolver

type DNSResolver interface {
	Start(<-chan struct{}) error
	GetDomain() string
	AddService(service string) (string, error)
	RemoveService(service string) error
	SyncServices(services map[string]bool) error
	ForwardLookup(name string) (string, error)
	ReverseLookup(ip string) (string, error)
}
