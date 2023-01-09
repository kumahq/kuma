package conntrack

type State string

const (
	// INVALID state means the packet is associated with no known connection
	INVALID State = "INVALID"
	// NEW state means the packet has started a new connection or otherwise
	// associated with a connection which has not seen packets in both
	// directions
	NEW State = "NEW"
	// ESTABLISHED state means the packet is associated with a connection which
	// has seen packets in both directions
	ESTABLISHED State = "ESTABLISHED"
	// RELATED state means the packet is starting a new connection, but
	// is associated with an existing connection, such as an FTP data transfer
	// or an ICMP error
	RELATED State = "RELATED"
	// UNTRACKED state means the packet is not tracked at all, which happens
	// if you explicitly untrack it by using -j CT --notrack in the raw table
	UNTRACKED State = "UNTRACKED"
	// SNAT state is a virtual state, matching if the original source address
	// differs from the reply destination
	SNAT State = "SNAT"
	// DNAT is a virtual state, matching if the original destination differs
	// from the reply source.
	DNAT State = "DNAT"
)
