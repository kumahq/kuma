package v1alpha1

const ZoneEgressServiceName = "zone-egress"

func (r *ZoneEgress) GetProxyType() ProxyTypeLabelValues {
	return ZoneEgressLabel
}
