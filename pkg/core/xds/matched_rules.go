package xds

import (
	"fmt"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

const ClientRuleAttachmentType = "clientSubset"
const DestinationRuleAttachmentType = "destinationSubset"
const SingleItemRuleAttachmentType = "singleItem"

type RuleAttachment struct {
	Type       string
	Name       string
	Service    string
	PolicyType core_model.ResourceType
	Rule       Rule
}

func BuildRulesAttachments(
	matchedPoliciesByType map[core_model.ResourceType]TypedMatchingPolicies,
	networking *mesh_proto.Dataplane_Networking,
) []RuleAttachment {
	var attachments []RuleAttachment

	for typ, matched := range matchedPoliciesByType {
		attachments = append(attachments, getInboundRuleAttachments(matched.FromRules.Rules, networking, typ)...)
		attachments = append(attachments, getOutboundRuleAttachments(matched.ToRules.Rules, networking, typ)...)
		if len(matched.SingleItemRules.Rules) > 0 {
			attachment := RuleAttachment{
				Type:       SingleItemRuleAttachmentType,
				Name:       "dataplane",
				PolicyType: typ,
				Rule:       *matched.SingleItemRules.Rules[0],
			}
			attachments = append(attachments, attachment)
		}
	}
	sort.SliceStable(attachments, func(i, j int) bool {
		if attachments[i].Name == attachments[j].Name {
			return attachments[i].Type < attachments[j].Type
		}
		return attachments[i].Name < attachments[j].Name
	})
	return attachments
}

func getInboundRuleAttachments(
	fromRules map[InboundListener]Rules,
	networking *mesh_proto.Dataplane_Networking,
	typ core_model.ResourceType,
) []RuleAttachment {
	inboundServices := map[InboundListener]string{}
	for _, inbound := range networking.GetInbound() {
		iface := networking.ToInboundInterface(inbound)
		inboundServices[InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}] = inbound.GetService()
	}

	var attachments []RuleAttachment
	for inbound, rules := range fromRules {
		for _, rule := range rules {
			attachment := RuleAttachment{
				Type:       ClientRuleAttachmentType,
				Name:       inbound.String(),
				Service:    inboundServices[inbound],
				PolicyType: typ,
				Rule:       *rule,
			}
			attachments = append(attachments, attachment)
		}
	}
	return attachments
}

func getOutboundRuleAttachments(
	rules Rules,
	networking *mesh_proto.Dataplane_Networking,
	typ core_model.ResourceType,
) []RuleAttachment {
	var attachments []RuleAttachment
	for _, outbound := range networking.Outbound {
		subset := SubsetFromTags(outbound.GetTagsIncludingLegacy())
		computedRule := rules.Compute(subset)
		if computedRule == nil {
			continue
		}
		oface := networking.ToOutboundInterface(outbound)
		attachment := RuleAttachment{
			Type:       DestinationRuleAttachmentType,
			Name:       fmt.Sprintf("%s:%d", oface.DataplaneIP, oface.DataplanePort),
			Service:    outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag],
			PolicyType: typ,
			Rule:       *computedRule,
		}
		attachments = append(attachments, attachment)
	}
	return attachments
}
