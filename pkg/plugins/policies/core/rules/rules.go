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
	"github.com/kumahq/kuma/pkg/core/kri"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
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
	// InboundRules is a map of InboundListener to a list of inbound rules built by using 'spec.rules' field.
	InboundRules map[InboundListener][]*inbound.Rule
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

	// OriginByMatches is an auxiliary structure for MeshHTTPRoute rules. It's a mapping between the rule (identified
	// by the hash of rule's matches) and the meta of the MeshHTTPRoute policy that contributed the rule.
	OriginByMatches map[common_api.MatchesHash]core_model.ResourceMeta
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
		rules, err := BuildRules(fromList, true)
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

func BuildToRules(matchedPolicies core_model.ResourceList, reader kri.ResourceReader) (ToRules, error) {
	rules, err := legacyBuildToRules(matchedPolicies, reader)
	if err != nil {
		return ToRules{}, err
	}

	// we have to exclude top-level targetRef 'MeshHTTPRoute' as new outbound rules work with MeshHTTPRoute differently,
	// see docs/madr/decisions/060-policy-matching-with-real-resources.md
	excludeTopLevelMeshHTTPRoute, err := registry.Global().NewList(matchedPolicies.GetItemType())
	if err != nil {
		return ToRules{}, err
	}
	for _, item := range matchedPolicies.GetItems() {
		if item.GetSpec().(core_model.Policy).GetTargetRef().Kind != common_api.MeshHTTPRoute {
			if err = excludeTopLevelMeshHTTPRoute.AddItem(item); err != nil {
				return ToRules{}, err
			}
		}
	}
	resourceRules, err := outbound.BuildRules(excludeTopLevelMeshHTTPRoute, reader)
	if err != nil {
		return ToRules{}, err
	}

	return ToRules{Rules: rules, ResourceRules: resourceRules}, nil
}

func legacyBuildToRules(matchedPolicies core_model.ResourceList, reader kri.ResourceReader) (Rules, error) {
	policiesWithTo, ok := common.Cast[core_model.PolicyWithToList](matchedPolicies.GetItems())
	if !ok {
		return Rules{}, nil
	}
	toList := []PolicyItemWithMeta{}
	for i, pwtl := range policiesWithTo {
		if idx := slices.IndexFunc(pwtl.GetToList(), func(item core_model.PolicyItem) bool {
			return item.GetTargetRef().Kind == common_api.MeshHTTPRoute
		}); idx >= 0 {
			continue
		}
		meta := matchedPolicies.GetItems()[i].GetMeta()
		tl, err := buildToListWithRoutes(meta, pwtl, reader.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).GetItems())
		if err != nil {
			return nil, err
		}
		if len(tl) > 0 {
			topLevel := pwtl.GetTargetRef()
			toList = append(toList, BuildPolicyItemsWithMeta(tl, meta, topLevel)...)
		}
	}
	return BuildRules(toList, false)
}

func BuildGatewayRules(
	matchedPoliciesByInbound map[InboundListener]core_model.ResourceList,
	matchedPoliciesByListener map[InboundListenerHostname]core_model.ResourceList,
	reader kri.ResourceReader,
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
		FromRules:    fromRules.Rules,
		InboundRules: fromRules.InboundRules,
	}, nil
}

func buildToListWithRoutes(meta core_model.ResourceMeta, policyWithTo core_model.PolicyWithToList, httpRoutes []core_model.Resource) ([]core_model.PolicyItem, error) {
	var mhr *meshhttproute_api.MeshHTTPRouteResource
	switch policyWithTo.GetTargetRef().Kind {
	case common_api.MeshHTTPRoute:
		for _, route := range httpRoutes {
			if core_model.IsReferenced(meta, pointer.Deref(policyWithTo.GetTargetRef().Name), route.GetMeta()) {
				if r, ok := route.(*meshhttproute_api.MeshHTTPRouteResource); ok {
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
	for _, mhrRules := range pointer.Deref(mhr.Spec.To) {
		for _, mhrRule := range mhrRules.Rules {
			matchesHash := meshhttproute_api.HashMatches(mhrRule.Matches)
			for _, to := range policyWithTo.GetToList() {
				var targetRef common_api.TargetRef
				switch mhrRules.TargetRef.Kind {
				case common_api.Mesh, common_api.MeshSubset:
					targetRef = common_api.TargetRef{
						Kind: common_api.MeshSubset,
						Tags: &map[string]string{
							RuleMatchesHashTag: string(matchesHash),
						},
					}
				default:
					targetRef = common_api.TargetRef{
						Kind: common_api.MeshServiceSubset,
						Name: mhrRules.TargetRef.Name,
						Tags: &map[string]string{
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

	rules, err := BuildRules(items, false)
	if err != nil {
		return SingleItemRules{}, err
	}

	return SingleItemRules{Rules: rules}, nil
}

// BuildRules creates a list of rules with negations sorted by the number of positive tags.
// If rules with negative tags are filtered out then the order becomes 'most specific to less specific'.
// Filtering out of negative rules could be useful for XDS generators that don't have a way to configure negations.
// In case of `to` policies we don't need to check negations since only possible value for `to` is either Mesh
// which has empty subset or kuma.io/service.
//
// See the detailed algorithm description in docs/madr/decisions/007-mesh-traffic-permission.md
func BuildRules(list []PolicyItemWithMeta, withNegations bool) (Rules, error) {
	rules := Rules{}
	oldKindsItems := []PolicyItemWithMeta{}
	for _, item := range list {
		if item.GetTargetRef().Kind.IsOldKind() {
			oldKindsItems = append(oldKindsItems, item)
		}
	}
	if len(oldKindsItems) == 0 {
		return rules, nil
	}

	uniqueKeys := map[string]struct{}{}
	// 1. Convert list of rules into the list of subsets
	var subsets []subsetutils.Subset
	for _, item := range oldKindsItems {
		ss, err := asSubset(item.GetTargetRef())
		if err != nil {
			return nil, err
		}
		for _, tag := range ss {
			uniqueKeys[tag.Key] = struct{}{}
		}
		subsets = append(subsets, ss)
	}

	// we don't need to generate all permutations when there is no negations
	// and we have only 0 or one tag, in other cases we need to generate.
	// in case of `to` policies it can happen when using top target ref MeshGateway,
	// for policy MeshHTTPRoute.
	if !withNegations && len(uniqueKeys) <= 1 {
		// deduplicate subsets
		subsets = subsetutils.Deduplicate(subsets)

		for _, ss := range subsets {
			if r, err := createRule(ss, oldKindsItems); err != nil {
				return nil, err
			} else {
				rules = append(rules, r...)
			}
		}

		sort.SliceStable(rules, func(i, j int) bool {
			// resource with more tags should be first
			return len(rules[i].Subset) > len(rules[j].Subset)
		})

		return rules, nil
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
			if r, err := createRule(ss, oldKindsItems); err != nil {
				return nil, err
			} else {
				rules = append(rules, r...)
			}
		}
	}

	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Subset.NumPositive() > rules[j].Subset.NumPositive()
	})

	return rules, nil
}

func createRule(ss subsetutils.Subset, items []PolicyItemWithMeta) ([]*Rule, error) {
	rules := []*Rule{}
	confs := []interface{}{}
	var relevant []PolicyItemWithMeta
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemSubset, err := asSubset(item.GetTargetRef())
		if err != nil {
			return nil, err
		}
		if itemSubset.IsSubset(ss) {
			confs = append(confs, item.GetDefault())
			relevant = append(relevant, item)
		}
	}

	getMeta := func(o common.Origin) core_model.ResourceMeta {
		return o.Resource
	}

	if len(relevant) > 0 {
		merged, err := merge.Confs(confs)
		if err != nil {
			return nil, err
		}
		for _, mergedRule := range merged {
			rules = append(rules, &Rule{
				Subset:          ss,
				Conf:            mergedRule,
				Origin:          util_slices.Map(common.Origins(relevant, false), getMeta),
				OriginByMatches: util_maps.MapValues(common.OriginByMatches(relevant), getMeta),
			})
		}
	}

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
		for k, v := range pointer.Deref(tr.Tags) {
			ss = append(ss, subsetutils.Tag{Key: k, Value: v})
		}
		return ss, nil
	case common_api.MeshService:
		return subsetutils.Subset{{Key: mesh_proto.ServiceTag, Value: pointer.Deref(tr.Name)}}, nil
	case common_api.MeshServiceSubset:
		ss := subsetutils.Subset{{Key: mesh_proto.ServiceTag, Value: pointer.Deref(tr.Name)}}
		for k, v := range pointer.Deref(tr.Tags) {
			ss = append(ss, subsetutils.Tag{Key: k, Value: v})
		}
		return ss, nil
	default:
		return nil, errors.Errorf("can't represent %s as tags", tr.Kind)
	}
}
