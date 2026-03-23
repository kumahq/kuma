package v1alpha1

const ZoneEgressServiceName = "zone-egress"

func (*ZoneEgress) GetProxyType() ProxyTypeLabelValues {
	return ZoneEgressLabel
}
