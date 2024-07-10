package policy

import (
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// ConnectionPolicy is a Policy that is applied on a connection between two data planes that match source and destination.
type ConnectionPolicy interface {
	core_model.Resource
	Sources() []*mesh_proto.Selector
	Destinations() []*mesh_proto.Selector
}

// OutboundConnectionPolicyMap holds the most specific ConnectionPolicy for each outbound interface of a Dataplane.
type OutboundConnectionPolicyMap map[core_xds.ServiceName]ConnectionPolicy

type InboundConnectionPolicyMap map[mesh_proto.InboundInterface]ConnectionPolicy

type InboundConnectionPoliciesMap map[mesh_proto.InboundInterface][]ConnectionPolicy

// DataplanePolicy is a Policy that is applied on a selected Dataplane
type DataplanePolicy interface {
	core_model.Resource
	Selectors() []*mesh_proto.Selector
}

type InboundDataplanePolicyMap map[mesh_proto.InboundInterface]DataplanePolicy

func MatchingRegex(tags mesh_proto.SingleValueTagSet) string {
	var re string
	for _, key := range tags.Keys() {
		keyIsEqual := fmt.Sprintf(`&%s=`, key)
		var value string
		switch tags[key] {
		case "*":
			value = ``
		default:
			value = fmt.Sprintf(`[^&]*%s[,&]`, tags[key])
		}
		value = strings.ReplaceAll(value, ".", `\.`)
		expr := keyIsEqual + value + `.*`
		re += expr
	}
	re = `.*` + re
	return re
}

func RegexOR(r ...string) string {
	if len(r) == 0 {
		return ""
	}
	if len(r) == 1 {
		return r[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(r, "|"))
}

func MatchSourceRegex(policy ConnectionPolicy) string {
	var selectorRegexs []string
	for _, selector := range policy.Sources() {
		selectorRegexs = append(selectorRegexs, MatchingRegex(selector.Match))
	}
	return RegexOR(selectorRegexs...)
}
