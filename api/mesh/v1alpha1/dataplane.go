package v1alpha1

// EnvoyAdmin describes the Envoy Admin API server of a proxy.
type EnvoyAdmin struct {
	// Port on which Envoy Admin API server will be listening.
	Port uint32 `json:"port,omitempty"`
}

func (a *EnvoyAdmin) GetPort() uint32 {
	if a == nil {
		return 0
	}
	return a.Port
}

// Dataplane defines a configuration of a side-car proxy.
type Dataplane struct {
	Networking *Dataplane_Networking `json:"networking,omitempty"`
}

func (d *Dataplane) GetNetworking() *Dataplane_Networking {
	if d == nil {
		return nil
	}
	return d.Networking
}

// Dataplane_Networking describes inbound and outbound interfaces of a proxy.
type Dataplane_Networking struct {
	// Address on which inbound listeners will be exposed.
	Address string `json:"address,omitempty"`
	// Inbound describes a service implemented by the proxy.
	Inbound []*Dataplane_Networking_Inbound `json:"inbound,omitempty"`
	// Outbound describes a service consumed by the proxy.
	Outbound []*Dataplane_Networking_Outbound `json:"outbound,omitempty"`
	// TransparentProxying describes the iptables based redirection.
	TransparentProxying *Dataplane_Networking_TransparentProxying `json:"transparentProxying,omitempty"`
	// Admin describes the Envoy Admin API of the proxy.
	Admin *EnvoyAdmin `json:"admin,omitempty"`
	// Listeners describes the zone proxy listeners of the proxy.
	Listeners []*Dataplane_Networking_Listener `json:"listeners,omitempty"`
}

func (n *Dataplane_Networking) GetAddress() string {
	if n == nil {
		return ""
	}
	return n.Address
}

func (n *Dataplane_Networking) GetInbound() []*Dataplane_Networking_Inbound {
	if n == nil {
		return nil
	}
	return n.Inbound
}

func (n *Dataplane_Networking) GetOutbound() []*Dataplane_Networking_Outbound {
	if n == nil {
		return nil
	}
	return n.Outbound
}

func (n *Dataplane_Networking) GetTransparentProxying() *Dataplane_Networking_TransparentProxying {
	if n == nil {
		return nil
	}
	return n.TransparentProxying
}

func (n *Dataplane_Networking) GetAdmin() *EnvoyAdmin {
	if n == nil {
		return nil
	}
	return n.Admin
}

func (n *Dataplane_Networking) GetListeners() []*Dataplane_Networking_Listener {
	if n == nil {
		return nil
	}
	return n.Listeners
}

// Dataplane_Networking_Inbound describes a service implemented by the proxy.
type Dataplane_Networking_Inbound struct {
	// Port of the inbound interface that will forward requests to the service.
	Port uint32 `json:"port,omitempty"`
	// Port of the service that requests will be forwarded to.
	ServicePort uint32 `json:"servicePort,omitempty"`
	// Address of the service that requests will be forwarded to.
	ServiceAddress string `json:"serviceAddress,omitempty"`
	// Address on which the inbound listener will be exposed.
	Address string `json:"address,omitempty"`
	// Health describes the health of the service.
	Health *Dataplane_Networking_Inbound_Health `json:"health,omitempty"`
	// ServiceProbe defines the way the proxy checks the service.
	ServiceProbe *Dataplane_Networking_Inbound_ServiceProbe `json:"serviceProbe,omitempty"`
	// State describes the current state of the inbound.
	State Dataplane_Networking_Inbound_State `json:"state,omitempty"`
	// Name adds another way of referencing this port, usable with MeshService.
	Name string `json:"name,omitempty"`
	// Protocol of the service.
	Protocol string `json:"protocol,omitempty"`
}

func (i *Dataplane_Networking_Inbound) GetPort() uint32 {
	if i == nil {
		return 0
	}
	return i.Port
}

func (i *Dataplane_Networking_Inbound) GetServicePort() uint32 {
	if i == nil {
		return 0
	}
	return i.ServicePort
}

func (i *Dataplane_Networking_Inbound) GetServiceAddress() string {
	if i == nil {
		return ""
	}
	return i.ServiceAddress
}

func (i *Dataplane_Networking_Inbound) GetAddress() string {
	if i == nil {
		return ""
	}
	return i.Address
}

func (i *Dataplane_Networking_Inbound) GetHealth() *Dataplane_Networking_Inbound_Health {
	if i == nil {
		return nil
	}
	return i.Health
}

func (i *Dataplane_Networking_Inbound) GetServiceProbe() *Dataplane_Networking_Inbound_ServiceProbe {
	if i == nil {
		return nil
	}
	return i.ServiceProbe
}

func (i *Dataplane_Networking_Inbound) GetState() Dataplane_Networking_Inbound_State {
	if i == nil {
		return Dataplane_Networking_Inbound_Ready
	}
	return i.State
}

func (i *Dataplane_Networking_Inbound) GetName() string {
	if i == nil {
		return ""
	}
	return i.Name
}

func (i *Dataplane_Networking_Inbound) GetProtocol() string {
	if i == nil {
		return ""
	}
	return i.Protocol
}

// Dataplane_Networking_Inbound_Health describes the health of a service.
type Dataplane_Networking_Inbound_Health struct {
	Ready bool `json:"ready,omitempty"`
}

func (h *Dataplane_Networking_Inbound_Health) GetReady() bool {
	if h == nil {
		return false
	}
	return h.Ready
}

// Dataplane_Networking_Inbound_ServiceProbe defines how the proxy checks the service.
type Dataplane_Networking_Inbound_ServiceProbe struct {
	Interval           *Duration                                      `json:"interval,omitempty"`
	Timeout            *Duration                                      `json:"timeout,omitempty"`
	UnhealthyThreshold *UInt32Value                                   `json:"unhealthyThreshold,omitempty"`
	HealthyThreshold   *UInt32Value                                   `json:"healthyThreshold,omitempty"`
	Tcp                *Dataplane_Networking_Inbound_ServiceProbe_Tcp `json:"tcp,omitempty"`
}

func (p *Dataplane_Networking_Inbound_ServiceProbe) GetInterval() *Duration {
	if p == nil {
		return nil
	}
	return p.Interval
}

func (p *Dataplane_Networking_Inbound_ServiceProbe) GetTimeout() *Duration {
	if p == nil {
		return nil
	}
	return p.Timeout
}

func (p *Dataplane_Networking_Inbound_ServiceProbe) GetUnhealthyThreshold() *UInt32Value {
	if p == nil {
		return nil
	}
	return p.UnhealthyThreshold
}

func (p *Dataplane_Networking_Inbound_ServiceProbe) GetHealthyThreshold() *UInt32Value {
	if p == nil {
		return nil
	}
	return p.HealthyThreshold
}

func (p *Dataplane_Networking_Inbound_ServiceProbe) GetTcp() *Dataplane_Networking_Inbound_ServiceProbe_Tcp {
	if p == nil {
		return nil
	}
	return p.Tcp
}

// Dataplane_Networking_Inbound_ServiceProbe_Tcp selects a plain TCP probe.
type Dataplane_Networking_Inbound_ServiceProbe_Tcp struct{}

// Dataplane_Networking_Outbound describes a service consumed by the proxy.
type Dataplane_Networking_Outbound struct {
	// Address on which the proxy will listen for the outbound traffic.
	Address string `json:"address,omitempty"`
	// Port on which the proxy will listen for the outbound traffic.
	Port uint32 `json:"port,omitempty"`
	// BackendRef is the destination the outbound resolves to.
	BackendRef *Dataplane_Networking_Outbound_BackendRef `json:"backendRef,omitempty"`
}

func (o *Dataplane_Networking_Outbound) GetAddress() string {
	if o == nil {
		return ""
	}
	return o.Address
}

func (o *Dataplane_Networking_Outbound) GetPort() uint32 {
	if o == nil {
		return 0
	}
	return o.Port
}

func (o *Dataplane_Networking_Outbound) GetBackendRef() *Dataplane_Networking_Outbound_BackendRef {
	if o == nil {
		return nil
	}
	return o.BackendRef
}

// Dataplane_Networking_Outbound_BackendRef points an outbound at its destination.
type Dataplane_Networking_Outbound_BackendRef struct {
	Kind   string            `json:"kind,omitempty"`
	Name   string            `json:"name,omitempty"`
	Port   uint32            `json:"port,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

func (r *Dataplane_Networking_Outbound_BackendRef) GetKind() string {
	if r == nil {
		return ""
	}
	return r.Kind
}

func (r *Dataplane_Networking_Outbound_BackendRef) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *Dataplane_Networking_Outbound_BackendRef) GetPort() uint32 {
	if r == nil {
		return 0
	}
	return r.Port
}

func (r *Dataplane_Networking_Outbound_BackendRef) GetLabels() map[string]string {
	if r == nil {
		return nil
	}
	return r.Labels
}

// Dataplane_Networking_TransparentProxying describes the iptables based redirection. The
// redirect ports and the IP family mode reach the control plane in kuma-dp's metadata
// rather than here.
type Dataplane_Networking_TransparentProxying struct {
	// List of services that will be accessed directly via IP:PORT.
	DirectAccessServices []string `json:"directAccessServices,omitempty"`
	// ReachableBackends limits the backends the proxy is configured for.
	ReachableBackends *Dataplane_Networking_TransparentProxying_ReachableBackends `json:"reachableBackends,omitempty"`
}

func (t *Dataplane_Networking_TransparentProxying) GetDirectAccessServices() []string {
	if t == nil {
		return nil
	}
	return t.DirectAccessServices
}

func (t *Dataplane_Networking_TransparentProxying) GetReachableBackends() *Dataplane_Networking_TransparentProxying_ReachableBackends {
	if t == nil {
		return nil
	}
	return t.ReachableBackends
}

// Dataplane_Networking_TransparentProxying_ReachableBackendRef names one reachable backend.
type Dataplane_Networking_TransparentProxying_ReachableBackendRef struct {
	Kind      string            `json:"kind,omitempty"`
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Port      *UInt32Value      `json:"port,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

func (r *Dataplane_Networking_TransparentProxying_ReachableBackendRef) GetKind() string {
	if r == nil {
		return ""
	}
	return r.Kind
}

func (r *Dataplane_Networking_TransparentProxying_ReachableBackendRef) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *Dataplane_Networking_TransparentProxying_ReachableBackendRef) GetNamespace() string {
	if r == nil {
		return ""
	}
	return r.Namespace
}

func (r *Dataplane_Networking_TransparentProxying_ReachableBackendRef) GetPort() *UInt32Value {
	if r == nil {
		return nil
	}
	return r.Port
}

func (r *Dataplane_Networking_TransparentProxying_ReachableBackendRef) GetLabels() map[string]string {
	if r == nil {
		return nil
	}
	return r.Labels
}

// Dataplane_Networking_TransparentProxying_ReachableBackends holds the reachable refs.
type Dataplane_Networking_TransparentProxying_ReachableBackends struct {
	Refs []*Dataplane_Networking_TransparentProxying_ReachableBackendRef `json:"refs,omitempty"`
}

func (b *Dataplane_Networking_TransparentProxying_ReachableBackends) GetRefs() []*Dataplane_Networking_TransparentProxying_ReachableBackendRef {
	if b == nil {
		return nil
	}
	return b.Refs
}

// Dataplane_Networking_Listener describes a zone proxy listener.
type Dataplane_Networking_Listener struct {
	// Type distinguishes a zone ingress listener from a zone egress one.
	Type Dataplane_Networking_Listener_Type `json:"type,omitempty"`
	// Address on which the listener is exposed.
	Address string `json:"address,omitempty"`
	// Port on which the listener is exposed.
	Port uint32 `json:"port,omitempty"`
	// Name of the listener.
	Name string `json:"name,omitempty"`
	// State describes the current state of the listener.
	State Dataplane_Networking_Listener_State `json:"state,omitempty"`
}

func (l *Dataplane_Networking_Listener) GetType() Dataplane_Networking_Listener_Type {
	if l == nil {
		return Dataplane_Networking_Listener_Unspecified
	}
	return l.Type
}

func (l *Dataplane_Networking_Listener) GetAddress() string {
	if l == nil {
		return ""
	}
	return l.Address
}

func (l *Dataplane_Networking_Listener) GetPort() uint32 {
	if l == nil {
		return 0
	}
	return l.Port
}

func (l *Dataplane_Networking_Listener) GetName() string {
	if l == nil {
		return ""
	}
	return l.Name
}

func (l *Dataplane_Networking_Listener) GetState() Dataplane_Networking_Listener_State {
	if l == nil {
		return Dataplane_Networking_Listener_Ready
	}
	return l.State
}
