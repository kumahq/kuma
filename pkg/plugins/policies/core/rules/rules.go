package rules

import (
	"encoding"
	"fmt"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
)

type InboundListener struct {
	Address string
	Port    uint32
}

// We need to implement TextMarshaler because InboundListener is used
// as a key for maps that are JSON encoded for logging.
var _ encoding.TextMarshaler = InboundListener{}

func (i InboundListener) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i InboundListener) String() string {
	return fmt.Sprintf("%s:%d", i.Address, i.Port)
}

type FromRules struct {
	// InboundRules is a map of InboundListener to a list of inbound rules built by using 'spec.rules' field.
	InboundRules map[InboundListener][]*inbound.Rule
}

type ToRules struct {
	ResourceRules outbound.ResourceRules
}

// ProxyConf is the single merged configuration of a proxy-wide policy applying
// to the whole proxy. A nil *ProxyConf means either no policy of that type matched
// the proxy, or the matched policy type doesn't support proxy-wide configuration.
type ProxyConf struct {
	Conf   any
	Origin []core_model.ResourceMeta
}

type Rule struct {
	Subset          subsetutils.Subset
	Conf            any
	Origin          []core_model.ResourceMeta
	OriginByMatches map[common_api.MatchesHash]core_model.ResourceMeta
}

type Rules []*Rule

type PolicyItemWithMeta struct {
	core_model.PolicyItem
	core_model.ResourceMeta
	TopLevel  common_api.TargetRef
	RuleIndex int
}

func (p PolicyItemWithMeta) GetTopLevel() common_api.TargetRef {
	return p.TopLevel
}

func (p PolicyItemWithMeta) GetResourceMeta() core_model.ResourceMeta {
	return p.ResourceMeta
}

func (p PolicyItemWithMeta) GetRuleIndex() int {
	return p.RuleIndex
}

func (p PolicyItemWithMeta) GetEntry() core_model.PolicyItem {
	return p.PolicyItem
}

func BuildFromRules(
	matchedPoliciesByInbound map[InboundListener]core_model.ResourceList,
) (FromRules, error) {
	rulesByInboundNew := map[InboundListener][]*inbound.Rule{}

	for inb, policies := range matchedPoliciesByInbound {
		rulesNew, err := inbound.BuildRules(policies)
		if err != nil {
			return FromRules{}, err
		}
		rulesByInboundNew[inb] = rulesNew
	}
	return FromRules{
		InboundRules: rulesByInboundNew,
	}, nil
}

func BuildToRules(matchedPolicies core_model.ResourceList, reader kri.ResourceReader) (ToRules, error) {
	resourceRules, err := outbound.BuildRules(matchedPolicies, reader)
	if err != nil {
		return ToRules{}, err
	}
	return ToRules{ResourceRules: resourceRules}, nil
}

func BuildProxyConf(matchedPolicies []core_model.Resource) (*ProxyConf, error) {
	if len(matchedPolicies) == 0 {
		return nil, nil
	}

	items := []PolicyItemWithMeta{}
	confs := []any{}
	for _, mp := range matchedPolicies {
		policyWithSingleItem, ok := mp.GetSpec().(core_model.PolicyWithSingleItem)
		if !ok {
			// policy doesn't support single item
			return nil, nil
		}
		item := PolicyItemWithMeta{
			PolicyItem:   policyWithSingleItem.GetPolicyItem(),
			ResourceMeta: mp.GetMeta(),
		}
		items = append(items, item)
		confs = append(confs, item.GetDefault())
	}

	merged, err := merge.Confs(confs)
	if err != nil {
		return nil, err
	}
	if len(merged) == 0 {
		return nil, nil
	}
	if len(merged) > 1 {
		return nil, errors.Errorf("expected a single merged proxy-wide config, got %d", len(merged))
	}

	return &ProxyConf{
		Conf:   merged[0],
		Origin: util_slices.Map(common.Origins(items, false), func(o common.Origin) core_model.ResourceMeta { return o.Resource }),
	}, nil
}
