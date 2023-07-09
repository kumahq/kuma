package inspect

import (
	"fmt"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

const (
	ClientRuleAttachmentType      = "ClientSubset"
	DestinationRuleAttachmentType = "DestinationSubset"
	SingleItemRuleAttachmentType  = "SingleItem"
)

type RuleAttachment struct {
	Type       string
	Name       string
	Addresses  []string
	Service    string
	Tags       map[string]string
	PolicyType core_model.ResourceType
	Rule       core_rules.Rule
}

func (r *RuleAttachment) AddAddress(address string) {
	for _, a := range r.Addresses {
		if a == address {
			return
		}
	}
	r.Addresses = append(r.Addresses, address)
}

func BuildRulesAttachments(matchedPoliciesByType map[core_model.ResourceType]core_xds.TypedMatchingPolicies, networking *mesh_proto.Dataplane_Networking, domains []core_xds.VIPDomains) []RuleAttachment {
	domainsByAddress := map[string][]string{}
	for _, d := range domains {
		domainsByAddress[d.Address] = append(domainsByAddress[d.Address], d.Domains...)
	}
	var attachments []RuleAttachment

	for typ, matched := range matchedPoliciesByType {
		attachments = append(attachments, getInboundRuleAttachments(matched.FromRules.Rules, networking, typ)...)
		attachments = append(attachments, getOutboundRuleAttachments(matched.ToRules.Rules, networking, typ, domainsByAddress)...)
		if len(matched.SingleItemRules.Rules) > 0 {
			attachment := RuleAttachment{
				Type:       SingleItemRuleAttachmentType,
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
	fromRules map[core_rules.InboundListener]core_rules.Rules,
	networking *mesh_proto.Dataplane_Networking,
	typ core_model.ResourceType,
) []RuleAttachment {
	inboundServices := map[core_rules.InboundListener]tags.Tags{}
	for _, inbound := range networking.GetInbound() {
		iface := networking.ToInboundInterface(inbound)
		inboundServices[core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}] = inbound.GetTags()
	}

	var attachments []RuleAttachment
	for inbound, rules := range fromRules {
		for _, rule := range rules {
			name, err := inboundServices[inbound].DestinationClusterName(nil)
			if err != nil {
				panic(err)
			}
			attachment := RuleAttachment{
				Type:       ClientRuleAttachmentType,
				Name:       name,
				Tags:       inboundServices[inbound],
				Addresses:  []string{inbound.String()},
				Service:    inboundServices[inbound][mesh_proto.ServiceTag],
				PolicyType: typ,
				Rule:       *rule,
			}
			attachments = append(attachments, attachment)
		}
	}
	return attachments
}

func getOutboundRuleAttachments(rules core_rules.Rules, networking *mesh_proto.Dataplane_Networking, typ core_model.ResourceType, domainsByAddress map[string][]string) []RuleAttachment {
	var attachments []RuleAttachment
	byUniqueClusterName := map[string]*RuleAttachment{}
	for _, outbound := range networking.Outbound {
		outboundTags := outbound.GetTags()
		name, err := tags.Tags(outboundTags).DestinationClusterName(nil)
		if err != nil {
			// Error is impossible here (there's always a service on Outbound)
			panic(err)
		}
		attachment := byUniqueClusterName[name]
		if attachment == nil {
			subset := core_rules.SubsetFromTags(outboundTags)
			computedRule := rules.Compute(subset)
			if computedRule == nil {
				continue
			}
			attachments = append(attachments, RuleAttachment{
				Name:       name,
				Type:       DestinationRuleAttachmentType,
				Service:    outbound.GetService(),
				Tags:       outboundTags,
				PolicyType: typ,
				Rule:       *computedRule,
			})
			attachment = &attachments[len(attachments)-1]
			byUniqueClusterName[name] = attachment
		}
		oface := networking.ToOutboundInterface(outbound)
		// reverse lookup address
		for _, d := range domainsByAddress[oface.DataplaneIP] {
			attachment.AddAddress(fmt.Sprintf("%s:%d", d, oface.DataplanePort))
		}
		// Add the ip anyway
		attachment.AddAddress(fmt.Sprintf("%s:%d", oface.DataplaneIP, oface.DataplanePort))
	}
	return attachments
}
