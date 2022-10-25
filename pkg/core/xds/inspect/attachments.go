package inspect

import (
	"fmt"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
)

type AttachmentType int64

const (
	Inbound AttachmentType = iota
	Outbound
	Service
	Dataplane
)

func (s AttachmentType) String() string {
	switch s {
	case Inbound:
		return "inbound"
	case Outbound:
		return "outbound"
	case Service:
		return "service"
	case Dataplane:
		return "dataplane"
	default:
		return "unknown"
	}
}

type Attachment struct {
	Type AttachmentType
	Name string
	// Service indicates service for the attachments.
	// For Dataplane AttachmentType it's empty since we are not matching to a specific service.
	Service string
}

type PolicyKey struct {
	Type core_model.ResourceType
	Key  core_model.ResourceKey
}

type PoliciesByResourceType map[core_model.ResourceType][]core_model.Resource

type AttachmentList []Attachment
type Attachments map[Attachment][]core_model.Resource

type AttachmentMap map[Attachment]PoliciesByResourceType
type AttachmentsByPolicy map[PolicyKey]AttachmentList

func (abp AttachmentsByPolicy) Merge(other AttachmentsByPolicy) {
	for k, v := range other {
		abp[k] = append(abp[k], v...)
	}
}

func (a AttachmentList) Len() int           { return len(a) }
func (a AttachmentList) Less(i, j int) bool { return fmt.Sprintf("%s", a[i]) < fmt.Sprintf("%s", a[j]) }
func (a AttachmentList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func BuildAttachments(matchedPolicies *xds.MatchedPolicies, networking *mesh_proto.Dataplane_Networking) Attachments {
	attachments := Attachments{}

	serviceByInbound := map[mesh_proto.InboundInterface]string{}
	for _, iface := range networking.GetInbound() {
		serviceByInbound[networking.ToInboundInterface(iface)] = iface.GetService()
	}

	for inbound, policies := range getInboundMatchedPolicies(matchedPolicies) {
		attachment := Attachment{
			Type:    Inbound,
			Name:    inbound.String(),
			Service: serviceByInbound[inbound],
		}
		attachments[attachment] = append(attachments[attachment], policies...)
	}

	serviceByOutbound := map[mesh_proto.OutboundInterface]string{}
	for _, oface := range networking.GetOutbound() {
		serviceByOutbound[networking.ToOutboundInterface(oface)] = oface.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
	}

	for outbound, policies := range getOutboundMatchedPolicies(matchedPolicies) {
		attachment := Attachment{
			Type:    Outbound,
			Name:    outbound.String(),
			Service: serviceByOutbound[outbound],
		}
		attachments[attachment] = append(attachments[attachment], policies...)
	}

	for service, policies := range getServiceMatchedPolicies(matchedPolicies) {
		attachment := Attachment{
			Type:    Service,
			Name:    service,
			Service: service,
		}
		attachments[attachment] = append(attachments[attachment], policies...)
	}

	attachments[Attachment{Type: Dataplane, Name: ""}] = getDataplaneMatchedPolicies(matchedPolicies)

	return attachments
}

func GroupByAttachment(matchedPolicies *xds.MatchedPolicies, networking *mesh_proto.Dataplane_Networking) AttachmentMap {
	result := AttachmentMap{}

	for attachment, policies := range BuildAttachments(matchedPolicies, networking) {
		if len(policies) == 0 {
			continue
		}
		if _, ok := result[attachment]; !ok {
			result[attachment] = PoliciesByResourceType{}
		}
		for _, policy := range policies {
			resType := policy.Descriptor().Name
			result[attachment][resType] = append(result[attachment][resType], policy)
		}
	}

	return result
}

func GroupByPolicy(matchedPolicies *xds.MatchedPolicies, networking *mesh_proto.Dataplane_Networking) AttachmentsByPolicy {
	result := AttachmentsByPolicy{}

	for attachment, policies := range BuildAttachments(matchedPolicies, networking) {
		for _, policy := range policies {
			key := PolicyKey{
				Type: policy.Descriptor().Name,
				Key:  core_model.MetaToResourceKey(policy.GetMeta()),
			}
			result[key] = append(result[key], attachment)
		}
	}

	for policyKey := range result {
		sort.Stable(result[policyKey])
	}

	return result
}

func getInboundMatchedPolicies(matchedPolicies *xds.MatchedPolicies) map[mesh_proto.InboundInterface][]core_model.Resource {
	result := map[mesh_proto.InboundInterface][]core_model.Resource{}

	for inbound, tp := range matchedPolicies.TrafficPermissions {
		result[inbound] = append(result[inbound], tp)
	}
	for inbound, fiList := range matchedPolicies.FaultInjections {
		for _, fi := range fiList {
			result[inbound] = append(result[inbound], fi)
		}
	}
	for inbound, rlList := range matchedPolicies.RateLimitsInbound {
		for _, rl := range rlList {
			result[inbound] = append(result[inbound], rl)
		}
	}
	for _, tpe := range matchedPolicies.OrderedDynamicPolicies() {
		for inbound, elts := range matchedPolicies.Dynamic[tpe].InboundPolicies {
			result[inbound] = append(result[inbound], elts...)
		}
	}

	return result
}

func getOutboundMatchedPolicies(matchedPolicies *xds.MatchedPolicies) map[mesh_proto.OutboundInterface][]core_model.Resource {
	result := map[mesh_proto.OutboundInterface][]core_model.Resource{}

	for outbound, timeout := range matchedPolicies.Timeouts {
		result[outbound] = append(result[outbound], timeout)
	}
	for outbound, rl := range matchedPolicies.RateLimitsOutbound {
		result[outbound] = append(result[outbound], rl)
	}
	for outbound, tr := range matchedPolicies.TrafficRoutes {
		result[outbound] = append(result[outbound], tr)
	}
	for _, tpe := range matchedPolicies.OrderedDynamicPolicies() {
		for outbound, elts := range matchedPolicies.Dynamic[tpe].OutboundPolicies {
			result[outbound] = append(result[outbound], elts...)
		}
	}

	return result
}

func getServiceMatchedPolicies(matchedPolicies *xds.MatchedPolicies) map[xds.ServiceName][]core_model.Resource {
	result := map[xds.ServiceName][]core_model.Resource{}

	for service, tl := range matchedPolicies.TrafficLogs {
		result[service] = append(result[service], tl)
	}
	for service, hc := range matchedPolicies.HealthChecks {
		result[service] = append(result[service], hc)
	}
	for service, cb := range matchedPolicies.CircuitBreakers {
		result[service] = append(result[service], cb)
	}
	for service, retry := range matchedPolicies.Retries {
		result[service] = append(result[service], retry)
	}
	for _, tpe := range matchedPolicies.OrderedDynamicPolicies() {
		for serviceName, elts := range matchedPolicies.Dynamic[tpe].ServicePolicies {
			result[serviceName] = append(result[serviceName], elts...)
		}
	}

	return result
}

func getDataplaneMatchedPolicies(matchedPolicies *xds.MatchedPolicies) []core_model.Resource {
	var resources []core_model.Resource
	if matchedPolicies.TrafficTrace != nil {
		resources = append(resources, matchedPolicies.TrafficTrace)
	}
	if matchedPolicies.ProxyTemplate != nil {
		resources = append(resources, matchedPolicies.ProxyTemplate)
	}
	for _, tpe := range matchedPolicies.OrderedDynamicPolicies() {
		resources = append(resources, matchedPolicies.Dynamic[tpe].DataplanePolicies...)
	}
	return resources
}
