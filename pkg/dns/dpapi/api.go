package dpapi

// API shared between Dataplane and Control Plane.

// PATH the path that this is exposed on the DP config endpoint
const PATH = "/dns"

type DNSRecord struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips"`
}

type DNSProxyConfig struct {
	Records []DNSRecord `json:"records"`
	TTL     int         `json:"ttl"`
}
