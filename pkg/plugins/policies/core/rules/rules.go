package rules

import (
	"encoding"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

const RuleMatchesHashTag = "__rule-matches-hash__"

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
	// Rules is a map of InboundListener to a list of rules built by using 'spec.from' field.
	// Deprecated: use InboundRules instead
	Rules map[InboundListener]Rules
	// InboundRules is a map of InboundListener to a list of inbound rules built by using 'spec.rules' field.
	InboundRules map[InboundListener][]*inbound.Rule
}

type ToRules struct {
	Rules         Rules
	ResourceRules outbound.ResourceRules
}

type InboundListenerHostname struct {
	Address  string
	Port     uint32
	hostname string
}

var _ encoding.TextMarshaler = InboundListenerHostname{}

func (i InboundListenerHostname) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i InboundListenerHostname) String() string {
	return fmt.Sprintf("%s:%d:%s", i.Address, i.Port, i.hostname)
}

func NewInboundListenerHostname(address string, port uint32, hostname string) InboundListenerHostname {
	if hostname == "" {
		hostname = mesh_proto.WildcardHostname
	}
	return InboundListenerHostname{
		Address:  address,
		Port:     port,
		hostname: hostname,
	}
}

type GatewayToRules struct {
	// ByListener contains rules that are not specific to hostnames
	// If the policy supports `GatewayListenerTagsAllowed: true`
	// then it likely should use ByListenerAndHostname
	ByListener map[InboundListener]ToRules
	// ByListenerAndHostname contains rules for policies that are specific to hostnames
	// This only relevant if the policy has `GatewayListenerTagsAllowed: true`
	ByListenerAndHostname map[InboundListenerHostname]ToRules
}

type GatewayRules struct {
	ToRules   GatewayToRules
	FromRules map[InboundListener]Rules
}

type SingleItemRules struct {
	Rules Rules
}

// Deprecated: use common.WithPolicyAttributes instead
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

func (p PolicyItemWithMeta) GetEntry() outbound.ToEntry {
	return p.PolicyItem
}

// Rule contains a configuration for the given Subset. When rule is an inbound rule (from),
// then Subset represents a group of clients. When rule is an outbound (to) then Subset
// represents destinations.
// Deprecated: use inbound.Rule or outbound.ResourceRule instead
type Rule struct {
	Subset subsetutils.Subset
	Conf   interface{}
	Origin []core_model.ResourceMeta

	// BackendRefOriginIndex is a mapping from the rule to the origin of the BackendRefs in the rule.
	// Some policies have BackendRefs in their confs, and it's important to know what was the original policy
	// that contributed the BackendRefs to the final conf. Rule (key) is represented as a hash from rule.Matches.
	// Origin (value) is represented as an index in the Origin list. If policy doesn't have rules (i.e. MeshTCPRoute)
	// then key is an empty string "".
	BackendRefOriginIndex common.BackendRefOriginIndex
}

func (r *Rule) GetBackendRefOrigin(hash common_api.MatchesHash) (core_model.ResourceMeta, bool) {
	if r == nil {
		return nil, false
	}
	if r.BackendRefOriginIndex == nil {
		return nil, false
	}
	index, ok := r.BackendRefOriginIndex[hash]
	if !ok {
		return nil, false
	}
	if index >= len(r.Origin) {
		return nil, false
	}
	return r.Origin[index], true
}

type Rules []*Rule

// Compute returns Rule for the given element.
func (rs Rules) Compute(element subsetutils.Element) *Rule {
	for _, rule := range rs {
		if rule.Subset.ContainsElement(element) {
			return rule
		}
	}
	return nil
}

// ComputeConf returns configuration for the given element.
func ComputeConf[T any](rs Rules, element subsetutils.Element) *T {
	computed := rs.Compute(element)
	if computed != nil {
		return pointer.To(computed.Conf.(T))
	}

	return nil
}

func BuildFromRules(
	matchedPoliciesByInbound map[InboundListener]core_model.ResourceList,
) (FromRules, error) {
	rulesByInbound := map[InboundListener]Rules{}
	rulesByInboundNew := map[InboundListener][]*inbound.Rule{}

	for inb, policies := range matchedPoliciesByInbound {
		fromList := []PolicyItemWithMeta{}
		for _, p := range policies.GetItems() {
			policyWithFrom, ok := p.GetSpec().(core_model.PolicyWithFromList)
			if !ok {
				return FromRules{}, nil
			}
			fromList = append(fromList, BuildPolicyItemsWithMeta(policyWithFrom.GetFromList(), p.GetMeta(), policyWithFrom.GetTargetRef())...)
		}
		rules, err := BuildRules(fromList)
		if err != nil {
			return FromRules{}, err
		}
		rulesByInbound[inb] = rules

		rulesNew, err := inbound.BuildRules(policies)
		if err != nil {
			return FromRules{}, err
		}
		rulesByInboundNew[inb] = rulesNew
	}
	return FromRules{
		Rules:        rulesByInbound,
		InboundRules: rulesByInboundNew,
	}, nil
}

func BuildToRules(matchedPolicies core_model.ResourceList, reader common.ResourceReader) (ToRules, error) {
	toList, err := buildToList(matchedPolicies.GetItems(), reader)
	if err != nil {
		return ToRules{}, err
	}

	rules, err := BuildRules(toList)
	if err != nil {
		return ToRules{}, err
	}

	// we have to exclude top-level targetRef 'MeshHTTPRoute' as new outbound rules work with MeshHTTPRoute differently,
	// see docs/madr/decisions/060-policy-matching-with-real-resources.md
	excludeTopLevelMeshHTTPRoute := slices.DeleteFunc(slices.Clone(matchedPolicies.GetItems()), func(r core_model.Resource) bool {
		return r.GetSpec().(core_model.Policy).GetTargetRef().Kind == common_api.MeshHTTPRoute
	})
	resourceRules, err := outbound.BuildRules(excludeTopLevelMeshHTTPRoute, reader)
	if err != nil {
		return ToRules{}, err
	}

	return ToRules{Rules: rules, ResourceRules: resourceRules}, nil
}

func buildToList(matchedPolicies []core_model.Resource, reader common.ResourceReader) ([]PolicyItemWithMeta, error) {
	toList := []PolicyItemWithMeta{}
	for _, mp := range matchedPolicies {
		tl, err := buildToListWithRoutes(mp, reader.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).GetItems())
		if err != nil {
			return nil, err
		}
		if len(tl) > 0 {
			topLevel := mp.GetSpec().(core_model.PolicyWithToList).GetTargetRef()
			toList = append(toList, BuildPolicyItemsWithMeta(tl, mp.GetMeta(), topLevel)...)
		}
	}
	return toList, nil
}

func BuildGatewayRules(
	matchedPoliciesByInbound map[InboundListener]core_model.ResourceList,
	matchedPoliciesByListener map[InboundListenerHostname]core_model.ResourceList,
	reader common.ResourceReader,
) (GatewayRules, error) {
	toRulesByInbound := map[InboundListener]ToRules{}
	toRulesByListenerHostname := map[InboundListenerHostname]ToRules{}
	for listener, policies := range matchedPoliciesByListener {
		toRules, err := BuildToRules(policies, reader)
		if err != nil {
			return GatewayRules{}, err
		}
		toRulesByListenerHostname[listener] = toRules
	}
	for inbound, policies := range matchedPoliciesByInbound {
		toRules, err := BuildToRules(policies, reader)
		if err != nil {
			return GatewayRules{}, err
		}
		toRulesByInbound[inbound] = toRules
	}

	fromRules, err := BuildFromRules(matchedPoliciesByInbound)
	if err != nil {
		return GatewayRules{}, err
	}

	return GatewayRules{
		ToRules: GatewayToRules{
			ByListenerAndHostname: toRulesByListenerHostname,
			ByListener:            toRulesByInbound,
		},
		FromRules: fromRules.Rules,
	}, nil
}

func buildToListWithRoutes(p core_model.Resource, httpRoutes []core_model.Resource) ([]core_model.PolicyItem, error) {
	policyWithTo, ok := p.GetSpec().(core_model.PolicyWithToList)
	if !ok {
		return nil, nil
	}

	var mhr *v1alpha1.MeshHTTPRouteResource
	switch policyWithTo.GetTargetRef().Kind {
	case common_api.MeshHTTPRoute:
		for _, route := range httpRoutes {
			if core_model.IsReferenced(p.GetMeta(), policyWithTo.GetTargetRef().Name, route.GetMeta()) {
				if r, ok := route.(*v1alpha1.MeshHTTPRouteResource); ok {
					mhr = r
				}
			}
		}
		if mhr == nil {
			return nil, errors.New("can't resolve MeshHTTPRoute policy")
		}
	default:
		return policyWithTo.GetToList(), nil
	}

	rv := []core_model.PolicyItem{}
	for _, mhrRules := range mhr.Spec.To {
		for _, mhrRule := range mhrRules.Rules {
			matchesHash := v1alpha1.HashMatches(mhrRule.Matches)
			for _, to := range policyWithTo.GetToList() {
				var targetRef common_api.TargetRef
				switch mhrRules.TargetRef.Kind {
				case common_api.Mesh, common_api.MeshSubset:
					targetRef = common_api.TargetRef{
						Kind: common_api.MeshSubset,
						Tags: map[string]string{
							RuleMatchesHashTag: string(matchesHash),
						},
					}
				default:
					targetRef = common_api.TargetRef{
						Kind: common_api.MeshServiceSubset,
						Name: mhrRules.TargetRef.Name,
						Tags: map[string]string{
							RuleMatchesHashTag: string(matchesHash),
						},
					}
				}
				rv = append(rv, &artificialPolicyItem{
					targetRef: targetRef,
					conf:      to.GetDefault(),
				})
			}
		}
	}

	return rv, nil
}

type artificialPolicyItem struct {
	conf      interface{}
	targetRef common_api.TargetRef
}

func (a *artificialPolicyItem) GetTargetRef() common_api.TargetRef {
	return a.targetRef
}

func (a *artificialPolicyItem) GetDefault() interface{} {
	return a.conf
}

func BuildPolicyItemsWithMeta(items []core_model.PolicyItem, meta core_model.ResourceMeta, topLevel common_api.TargetRef) []PolicyItemWithMeta {
	var result []PolicyItemWithMeta
	for i, item := range items {
		result = append(result, PolicyItemWithMeta{
			PolicyItem:   item,
			ResourceMeta: meta,
			TopLevel:     topLevel,
			RuleIndex:    i,
		})
	}
	return result
}

func BuildSingleItemRules(matchedPolicies []core_model.Resource) (SingleItemRules, error) {
	items := []PolicyItemWithMeta{}
	for _, mp := range matchedPolicies {
		policyWithSingleItem, ok := mp.GetSpec().(core_model.PolicyWithSingleItem)
		if !ok {
			// policy doesn't support single item
			return SingleItemRules{}, nil
		}
		item := PolicyItemWithMeta{
			PolicyItem:   policyWithSingleItem.GetPolicyItem(),
			ResourceMeta: mp.GetMeta(),
		}
		items = append(items, item)
	}

	rules, err := BuildRules(items)
	if err != nil {
		return SingleItemRules{}, err
	}

	return SingleItemRules{Rules: rules}, nil
}

// BuildRules creates a list of rules with negations sorted by the number of positive tags.
// If rules with negative tags are filtered out then the order becomes 'most specific to less specific'.
// Filtering out of negative rules could be useful for XDS generators that don't have a way to configure negations.
//
// See the detailed algorithm description in docs/madr/decisions/007-mesh-traffic-permission.md
func BuildRules(list []PolicyItemWithMeta) (Rules, error) {
	rules := Rules{}
	oldKindsItems := []PolicyItemWithMeta{}
	for _, item := range list {
		if item.PolicyItem.GetTargetRef().Kind.IsOldKind() {
			oldKindsItems = append(oldKindsItems, item)
		}
	}
	if len(oldKindsItems) == 0 {
		return rules, nil
	}

	// 1. Convert list of rules into the list of subsets
	var subsets []subsetutils.Subset
	for _, item := range oldKindsItems {
		ss, err := asSubset(item.GetTargetRef())
		if err != nil {
			return nil, err
		}
		subsets = append(subsets, ss)
	}

	// 2. Create a graph where nodes are subsets and edge exists between 2 subsets only if there is an intersection
	g := simple.NewUndirectedGraph()

	for nodeId := range subsets {
		g.AddNode(simple.Node(nodeId))
	}

	for i := range subsets {
		for j := range subsets {
			if i == j {
				continue
			}
			if subsets[i].Intersect(subsets[j]) {
				g.SetEdge(simple.Edge{F: simple.Node(i), T: simple.Node(j)})
			}
		}
	}

	// 3. Construct rules for all connected components of the graph independently
	components := topo.ConnectedComponents(g)

	sortComponents(components)

	for _, nodes := range components {
		tagSet := map[subsetutils.Tag]bool{}
		for _, node := range nodes {
			for _, t := range subsets[node.ID()] {
				tagSet[t] = true
			}
		}

		tags := []subsetutils.Tag{}
		for tag := range tagSet {
			tags = append(tags, tag)
		}

		sort.Slice(tags, func(i, j int) bool {
			if tags[i].Key != tags[j].Key {
				return tags[i].Key < tags[j].Key
			}
			return tags[i].Value < tags[j].Value
		})

		// 4. Iterate over all possible combinations with negations
		iter := subsetutils.NewSubsetIter(tags)
		for {
			ss := iter.Next()
			if ss == nil {
				break
			}
			// 5. For each combination determine a configuration
			confs := []interface{}{}
			var relevant []PolicyItemWithMeta
			for i := 0; i < len(oldKindsItems); i++ {
				item := oldKindsItems[i]
				itemSubset, err := asSubset(item.GetTargetRef())
				if err != nil {
					return nil, err
				}
				if itemSubset.IsSubset(ss) {
					confs = append(confs, item.GetDefault())
					relevant = append(relevant, item)
				}
			}

			if len(relevant) > 0 {
				merged, err := merge.Confs(confs)
				if err != nil {
					return nil, err
				}
				ruleOrigins, originIndex := common.Origins(relevant, false)
				resourceMetas := make([]core_model.ResourceMeta, 0, len(ruleOrigins))
				for _, o := range ruleOrigins {
					resourceMetas = append(resourceMetas, o.Resource)
				}
				for _, mergedRule := range merged {
					rules = append(rules, &Rule{
						Subset:                ss,
						Conf:                  mergedRule,
						Origin:                resourceMetas,
						BackendRefOriginIndex: originIndex,
					})
				}
			}
		}
	}

	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Subset.NumPositive() > rules[j].Subset.NumPositive()
	})

	return rules, nil
}

func sortComponents(components [][]graph.Node) {
	for _, c := range components {
		sort.SliceStable(c, func(i, j int) bool {
			return c[i].ID() < c[j].ID()
		})
	}
	sort.SliceStable(components, func(i, j int) bool {
		return strings.Join(toStringList(components[i]), ":") > strings.Join(toStringList(components[j]), ":")
	})
}

func toStringList(nodes []graph.Node) []string {
	rv := make([]string, 0, len(nodes))
	for _, id := range nodes {
		rv = append(rv, fmt.Sprintf("%d", id.ID()))
	}
	return rv
}

func asSubset(tr common_api.TargetRef) (subsetutils.Subset, error) {
	switch tr.Kind {
	case common_api.Mesh:
		return subsetutils.Subset{}, nil
	case common_api.MeshSubset:
		ss := subsetutils.Subset{}
		for k, v := range tr.Tags {
			ss = append(ss, subsetutils.Tag{Key: k, Value: v})
		}
		return ss, nil
	case common_api.MeshService:
		return subsetutils.Subset{{Key: mesh_proto.ServiceTag, Value: tr.Name}}, nil
	case common_api.MeshServiceSubset:
		ss := subsetutils.Subset{{Key: mesh_proto.ServiceTag, Value: tr.Name}}
		for k, v := range tr.Tags {
			ss = append(ss, subsetutils.Tag{Key: k, Value: v})
		}
		return ss, nil
	default:
		return nil, errors.Errorf("can't represent %s as tags", tr.Kind)
	}
}
