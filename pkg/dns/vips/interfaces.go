package vips

type List map[string]string

func (vips List) Append(other List) {
	for k, v := range other {
		vips[k] = v
	}
}

func (vips List) FQDNsByIPs() map[string]string {
	ipToDomain := map[string]string{}
	for domain, ip := range vips {
		ipToDomain[ip] = domain
	}
	return ipToDomain
}
