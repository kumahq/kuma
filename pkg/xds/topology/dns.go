package topology

import (
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
)

func Outbounds[T interface{ AsOutbounds() xds_types.Outbounds }](list []T) xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, item := range list {
		outbounds = append(outbounds, item.AsOutbounds()...)
	}
	return outbounds
}

func Domains[T interface{ Domains() *xds_types.VIPDomains }](list []T) []xds_types.VIPDomains {
	var domains []xds_types.VIPDomains
	for _, item := range list {
		if d := item.Domains(); d != nil {
			domains = append(domains, *d)
		}
	}
	return domains
}
