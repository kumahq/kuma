package resolver

type DNSResolver interface {
	Start(<-chan struct{}) error
	NeedLeaderElection() bool

	GetDomain() string
	AddService(service string) (string, error)
	RemoveService(service string) error
	SyncServices(services map[string]bool) error
	ForwardLookup(service string) (string, error)
	ForwardLookupFQDN(name string) (string, error)
	ReverseLookup(ip string) (string, error)
	SetElectedLeader(elected bool)
}
