package generator

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/dns/dpapi"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	util_net "github.com/kumahq/kuma/v3/pkg/util/net"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/dynconf"
)

type DNSGenerator struct{}

func (g DNSGenerator) Generate(_ context.Context, rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	tp := proxy.GetTransparentProxy()

	// DNS only makes sense when transparent proxy is used
	if !tp.Enabled() || !tp.Redirect.DNS.Enabled || proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) {
		return nil, nil
	}
	outboundIPs := map[string]bool{}
	for _, outbound := range proxy.Outbounds {
		outboundIPs[outbound.GetAddress()] = true
	}

	vips := map[string][]string{}
	for _, dnsOutbound := range xdsCtx.Mesh.VIPDomains {
		if !outboundIPs[dnsOutbound.Address] {
			continue // if there is no outbound for given address, there is no point of providing DNS resolver
		}
		addresses := []string{dnsOutbound.Address}
		v6 := util_net.ToV6(dnsOutbound.Address)
		if v6 != dnsOutbound.Address && tp.EnabledIPv6() {
			addresses = append(addresses, v6)
		}

		for _, domain := range dnsOutbound.Domains {
			vips[domain] = addresses
		}
	}
	// This is purposefully set to 30s to avoid DNS cache stale with ExternalService and Kong Gateway see: https://github.com/kumahq/kuma/issues/13353.
	// https://github.com/kumahq/kuma/issues/13463
	dnsInfo := dpapi.DNSProxyConfig{TTL: 30, Records: []dpapi.DNSRecord{}, ExtraLabels: dnsExtraLabels(proxy)}
	for name, addresses := range vips {
		dnsInfo.Records = append(dnsInfo.Records, dpapi.DNSRecord{
			Name: name,
			IPs:  addresses,
		})
	}

	slices.SortFunc(dnsInfo.Records, func(a, b dpapi.DNSRecord) int {
		return cmp.Compare(a.Name, b.Name)
	})
	bytes, err := json.Marshal(dnsInfo)
	if err != nil {
		return nil, err
	}
	if err := dynconf.AddConfigRoute(proxy, rs, true, "dns", dpapi.PATH, bytes); err != nil {
		return nil, err
	}
	return nil, nil
}

func dnsExtraLabels(proxy *core_xds.Proxy) map[string]string {
	labels := map[string]string{}
	dpLabels := proxy.Dataplane.GetMeta().GetLabels()

	if workloadName := dpLabels[k8s_metadata.KumaWorkload]; workloadName != "" {
		labels["kuma_workload"] = workloadName
	}
	if ns := dpLabels[mesh_proto.KubeNamespaceTag]; ns != "" {
		labels["k8s_kuma_io_namespace"] = ns
	}
	labels["mesh"] = proxy.Dataplane.GetMeta().GetMesh()
	if zone := dpLabels[mesh_proto.ZoneTag]; zone != "" {
		labels["kuma_io_zone"] = zone
	}
	return labels
}
