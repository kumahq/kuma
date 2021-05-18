package dns

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	util_net "github.com/kumahq/kuma/pkg/util/net"

	"github.com/kumahq/kuma/pkg/core"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/dns/vips"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var dnsOutboundGeneratorLog = core.Log.WithName("dns-outbound-generator")

const VIPListenPort = uint32(80)

func VIPOutbounds(
	resourceKey model.ResourceKey,
	dataplanes []*core_mesh.DataplaneResource,
	vips vips.List,
	externalServices []*core_mesh.ExternalServiceResource,
) []*mesh_proto.Dataplane_Networking_Outbound {
	if len(vips) == 0 {
		dnsOutboundGeneratorLog.Info("Skipping legacy generation as there are no vips")
	}
	type vipEntry struct {
		ip   string
		port uint32
	}
	serviceVIPMap := map[string]vipEntry{}
	services := []string{}
	for _, dataplane := range dataplanes {
		if dataplane.Spec.IsIngress() {
			for _, service := range dataplane.Spec.Networking.Ingress.AvailableServices {
				if service.Mesh == resourceKey.Mesh {
					// Only add outbounds for services in the same mesh
					inService := service.Tags[mesh_proto.ServiceTag]
					if _, found := serviceVIPMap[inService]; !found {
						vip, err := ForwardLookup(vips, inService)
						if err == nil {
							serviceVIPMap[inService] = vipEntry{vip, VIPListenPort}
							services = append(services, inService)
						}
					}
				}
			}
		} else {
			for _, inbound := range dataplane.Spec.Networking.Inbound {
				inService := inbound.GetTags()[mesh_proto.ServiceTag]
				if _, found := serviceVIPMap[inService]; !found {
					vip, err := ForwardLookup(vips, inService)
					if err == nil {
						serviceVIPMap[inService] = vipEntry{vip, VIPListenPort}
						services = append(services, inService)
					}
				}
			}
		}
	}

	for _, externalService := range externalServices {
		inService := externalService.Spec.Tags[mesh_proto.ServiceTag]
		if _, found := serviceVIPMap[inService]; !found {
			vip, err := ForwardLookup(vips, inService)
			if err == nil {
				port := externalService.Spec.GetPort()
				var p32 uint32
				if p64, err := strconv.ParseUint(port, 10, 32); err != nil {
					p32 = VIPListenPort
				} else {
					p32 = uint32(p64)
				}
				serviceVIPMap[inService] = vipEntry{vip, p32}
				services = append(services, inService)
			}
		}
	}

	sort.Strings(services)
	outbounds := []*mesh_proto.Dataplane_Networking_Outbound{}
	for _, service := range services {
		entry := serviceVIPMap[service]
		outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
			Address: entry.ip,
			Port:    entry.port,
			Tags:    map[string]string{mesh_proto.ServiceTag: service},
		})

		// todo (lobkovilya): backwards compatibility, could be deleted in the next major release Kuma 1.2.x
		if entry.port != VIPListenPort {
			outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
				Address: entry.ip,
				Port:    VIPListenPort,
				Tags:    map[string]string{mesh_proto.ServiceTag: service},
			})
		}
	}

	return outbounds
}

func ForwardLookup(vips vips.List, service string) (string, error) {
	ip, found := vips[service]
	if !found {
		return "", errors.Errorf("service [%s] not found", service)
	}
	return ip, nil
}

func VirtualOutbounds(
	resourceKey model.ResourceKey,
	dataplanes []*core_mesh.DataplaneResource,
	externalServices []*core_mesh.ExternalServiceResource,
	virtualOutbounds []*core_mesh.VirtualOutboundResource,
	cidr string,
) ([]*mesh_proto.Dataplane_Networking_Outbound, error) {
	self, tags := buildUniqueTagsList(resourceKey, dataplanes, externalServices)
	outbounds := buildOutbounds(tags, virtualOutbounds)

	ipam, err := NewSimpleIPAM(cidr)
	if err != nil {
		return nil, err
	}
	cidrIsV4 := util_net.CidrIsIpV4(cidr)
	ipByHostname := map[string]string{}
	if self != nil {
		// Retrieve existing ips from self to not change already assigned ips
		for _, outbound := range self.Spec.Networking.Outbound {
			if outbound.Hostname != "" && (!cidrIsV4 || util_net.IsV4(outbound.Address)) { // If we have a v4 cidr we only match things for v4
				err := ipam.ReserveIP(outbound.Address)
				if err != nil && !IsAddressAlreadyAllocated(err) && !IsAddressOutsideCidr(err) {
					return nil, errors.Wrapf(err, "Failed reserving ip: %s", outbound.Address)
				}
				ipByHostname[outbound.Hostname] = outbound.Address
			}
		}
	}
	var outboundWithv6 []*mesh_proto.Dataplane_Networking_Outbound
	for _, outbound := range outbounds {
		if _, ok := ipByHostname[outbound.Hostname]; !ok {
			// Allocate ip for hostname
			ip, err := ipam.AllocateIP()
			if err != nil {
				return nil, errors.Wrapf(err, "Failed allocating ip")
			}
			ipByHostname[outbound.Hostname] = ip
		}
		// Set the address for the hostname
		outbound.Address = ipByHostname[outbound.Hostname]
		outboundWithv6 = append(outboundWithv6, outbound)
		// Add a v6 listener is the ip wasn't v6
		if cidrIsV4 {
			outboundWithv6 = append(outboundWithv6, &mesh_proto.Dataplane_Networking_Outbound{
				Tags:     outbound.Tags,
				Port:     outbound.Port,
				Hostname: outbound.Hostname,
				Address:  util_net.ToV6(outbound.Address),
			})
		}
	}
	return outboundWithv6, nil
}

func buildUniqueTagsList(resourceKey model.ResourceKey, dataplanes []*core_mesh.DataplaneResource, externalServices []*core_mesh.ExternalServiceResource) (*core_mesh.DataplaneResource, []map[string]string) {
	tagSets := map[string]map[string]string{}
	allKeys := []string{}
	var self *core_mesh.DataplaneResource
	for _, dataplane := range dataplanes {
		if model.MetaToResourceKey(dataplane.Meta) == resourceKey {
			self = dataplane
		}
		if dataplane.Spec.IsIngress() {
			for _, inbound := range dataplane.Spec.Networking.Ingress.AvailableServices {
				if inbound.Mesh == resourceKey.Mesh {
					k := keyForTags(inbound.Tags)
					if tagSets[k] == nil {
						tagSets[k] = inbound.Tags
						allKeys = append(allKeys, k)
					}
				}
			}
		} else {
			for _, inbound := range dataplane.Spec.Networking.Inbound {
				k := keyForTags(inbound.Tags)
				if tagSets[k] == nil {
					tagSets[k] = inbound.Tags
					allKeys = append(allKeys, k)
				}
			}
		}
	}
	for _, externalService := range externalServices {
		k := keyForTags(externalService.Spec.Tags)
		if tagSets[k] == nil {
			tagSets[k] = externalService.Spec.Tags
			allKeys = append(allKeys, k)
		}
	}
	// Order is important for determinism
	sort.Strings(allKeys)
	tags := make([]map[string]string, len(allKeys))
	for i := range allKeys {
		tags[i] = tagSets[allKeys[i]]
	}
	return self, tags
}

func buildOutbounds(tagSets []map[string]string, policies []*core_mesh.VirtualOutboundResource) []*mesh_proto.Dataplane_Networking_Outbound {
	uniqueHostPort := map[string]*scoredOutbound{}
	for _, tagSet := range tagSets {
		for _, curPolicy := range policies {
			var bestRank *mesh_proto.TagSelectorRank
			if len(curPolicy.Selectors()) == 0 {
				bestRank = &mesh_proto.TagSelectorRank{}
			} else {
				for _, selector := range curPolicy.Selectors() {
					if len(selector.Match) == 0 {
						bestRank = &mesh_proto.TagSelectorRank{}
					} else {
						tagSelector := mesh_proto.TagSelector(selector.Match)
						if tagSelector.Matches(tagSet) {
							r := tagSelector.Rank()
							if bestRank == nil || r.CompareTo(*bestRank) > 0 {
								bestRank = &r
							}
						}
					}
				}
			}
			if bestRank != nil {
				r, err := newScoredOutbound(tagSet, *bestRank, curPolicy)
				if err != nil {
					dnsOutboundGeneratorLog.Error(err, "failed generating outbound", "tagSet", tagSet, "policy", curPolicy.GetMeta().GetName())
				} else {
					cur := uniqueHostPort[r.hostPort]
					if cur == nil || cur.Less(r) {
						uniqueHostPort[r.hostPort] = r
					}
				}
			}
		}
	}
	rankedOutbounds := make([]*scoredOutbound, 0, len(uniqueHostPort))
	for _, v := range uniqueHostPort {
		rankedOutbounds = append(rankedOutbounds, v)
	}
	sort.SliceStable(rankedOutbounds, func(i, j int) bool {
		return rankedOutbounds[i].Less(rankedOutbounds[j])
	})

	res := make([]*mesh_proto.Dataplane_Networking_Outbound, len(rankedOutbounds))
	for i := range rankedOutbounds {
		res[i] = rankedOutbounds[i].outbound
	}
	return res
}

type scoredOutbound struct {
	tagSets  map[string]string
	key      string
	hostPort string
	rank     mesh_proto.TagSelectorRank
	policy   *core_mesh.VirtualOutboundResource
	outbound *mesh_proto.Dataplane_Networking_Outbound
}

// Less by match inverse quality, then key
func (s *scoredOutbound) Less(other *scoredOutbound) bool {
	r := s.rank.CompareTo(other.rank)
	if r == 0 {
		if s.key == other.key {
			return s.hostPort < other.hostPort
		}
		return s.key < other.key
	}
	return r < 0
}

func newScoredOutbound(tagSet map[string]string, rank mesh_proto.TagSelectorRank, policy *core_mesh.VirtualOutboundResource) (*scoredOutbound, error) {
	filteredTags := policy.FilterTags(tagSet)
	s := &scoredOutbound{
		tagSets: tagSet,
		rank:    rank,
		policy:  policy,
		key:     fmt.Sprintf("%s{%s}", policy.Meta.GetName(), keyForTags(filteredTags)),
	}
	host, err := s.policy.EvalHost(s.tagSets)
	if err != nil {
		return nil, err
	}
	port, err := s.policy.EvalPort(s.tagSets)
	if err != nil {
		return nil, err
	}
	s.outbound = &mesh_proto.Dataplane_Networking_Outbound{
		Port:     port,
		Hostname: host,
		Tags:     filteredTags,
	}
	s.hostPort = fmt.Sprintf("%s:%d", s.outbound.Hostname, s.outbound.Port)
	return s, nil
}

func keyForTags(tags map[string]string) string {
	var t []string
	for k, v := range tags {
		t = append(t, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(t)
	return strings.Join(t, ",")
}
