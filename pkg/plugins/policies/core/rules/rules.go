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

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
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
	// InboundRules is a map of InboundListener to a list of inbound rules built by using 'spec.rules' field.
	InboundRules map[InboundListener][]*inbound.Rule
}

type ToRules struct {
	Rules         Rules
	ResourceRules outbound.ResourceRules
}

type MergedPolicyConf struct {
	Conf   any
	Origin []core_model.ResourceMeta
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
//
// Deprecated: use inbound.Rule or outbound.ResourceRule instead
type Rule struct {
	Subset subsetutils.Subset
	Conf   any
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
	rules, err := legacyBuildToRules(matchedPolicies, reader)
	if err != nil {
		return ToRules{}, err
	}

	// we have to exclude top-level targetRef 'MeshHTTPRoute' as new outbound rules work with MeshHTTPRoute differently,
	// see docs/madr/decisions/066-policy-matching-with-real-resources.md
	excludeTopLevelMeshHTTPRoute, err := registry.Global().NewList(matchedPolicies.GetItemType())
	if err != nil {
		return ToRules{}, err
	}
	for _, item := range matchedPolicies.GetItems() {
		if item.GetSpec().(core_model.Policy).GetTargetRef().Kind != common_api.MeshHTTPRoute {
			if err := excludeTopLevelMeshHTTPRoute.AddItem(item); err != nil {
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

func buildToListWithRoutes(meta core_model.ResourceMeta, policyWithTo core_model.PolicyWithToList, httpRoutes []core_model.Resource) ([]core_model.PolicyItem, error) {
	var mhrs []*meshhttproute_api.MeshHTTPRouteResource
	switch policyWithTo.GetTargetRef().Kind {
	case common_api.MeshHTTPRoute:
		routeLabels := pointer.Deref(policyWithTo.GetTargetRef().Labels)
		if len(routeLabels) == 0 {
			return nil, errors.New("can't resolve MeshHTTPRoute policy")
		}
		targetSubset := subsetutils.NewSubset(routeLabels)
		for _, route := range httpRoutes {
			if !targetSubset.IsSubset(subsetutils.NewSubset(route.GetMeta().GetLabels())) {
				continue
			}
			// A label subset can match routes that share the same labels
			// (e.g. kuma.io/display-name) across namespaces. Namespaced
			// policies must only expand routes from their own namespace,
			// otherwise route-derived config leaks across namespaces.
			if !policySelectsByNamespace(meta, route.GetMeta()) {
				continue
			}
			if r, ok := route.(*meshhttproute_api.MeshHTTPRouteResource); ok {
				mhrs = append(mhrs, r)
			}
		}
		if len(mhrs) == 0 {
			return nil, errors.New("can't resolve MeshHTTPRoute policy")
		}
	default:
		return policyWithTo.GetToList(), nil
	}

	rv := []core_model.PolicyItem{}
	for _, mhr := range mhrs {
		for _, mhrRules := range pointer.Deref(mhr.Spec.To) {
			for _, mhrRule := range mhrRules.Rules {
				matchesHash := meshhttproute_api.HashMatches(mhrRule.Matches)
				for _, to := range policyWithTo.GetToList() {
					var targetRef common_api.TargetRef
					switch mhrRules.TargetRef.Kind {
					case common_api.Mesh, common_api.LegacyMeshSubsetKind():
						targetRef = common_api.TargetRef{
							Kind: common_api.LegacyMeshSubsetKind(),
							Tags: &map[string]string{
								RuleMatchesHashTag: string(matchesHash),
							},
						}
					default:
						// The legacy subset keys on kuma.io/service, so the
						// route target must resolve to a service name. A label
						// selector that doesn't carry kuma.io/display-name has
						// no legacy equivalent; fail loudly rather than emitting
						// an empty selector that silently matches nothing.
						service := serviceTagValue(mhrRules.TargetRef.ToTargetRef())
						if service == "" {
							return nil, errors.Errorf("can't resolve %s targetRef to a service: kuma.io/display-name label is required", mhrRules.TargetRef.Kind)
						}
						targetRef = common_api.TargetRef{
							Kind: common_api.LegacyMeshServiceSubsetKind(),
							Labels: &map[string]string{
								mesh_proto.DisplayName: service,
							},
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
	}

	return rv, nil
}

type artificialPolicyItem struct {
	conf      any
	targetRef common_api.TargetRef
}

func (a *artificialPolicyItem) GetTargetRef() common_api.TargetRef {
	return a.targetRef
}

func (a *artificialPolicyItem) GetDefault() any {
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

func BuildMergedPolicyConf(matchedPolicies []core_model.Resource) (*MergedPolicyConf, error) {
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

	return &MergedPolicyConf{
		Conf:   merged[0],
		Origin: util_slices.Map(common.Origins(items, false), func(o common.Origin) core_model.ResourceMeta { return o.Resource }),
	}, nil
}

// BuildRules creates a list of rules with negations sorted by the number of positive tags.
// If rules with negative tags are filtered out then the order becomes 'most specific to less specific'.
// Filtering out of negative rules could be useful for XDS generators that don't have a way to configure negations.
// In case of `to` policies we don't need to check negations since only possible value for `to` is either Mesh
// which has empty subset or kuma.io/service.
//
// See the detailed algorithm description in docs/madr/decisions/007-mesh-traffic-permission.md
func BuildRules(list []PolicyItemWithMeta, withNegations bool) (Rules, error) {
	return buildRulesInternal(list, withNegations, true)
}

func buildRulesInternal(list []PolicyItemWithMeta, withNegations bool, useCliques bool) (Rules, error) {
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

	var nodeGroups [][]graph.Node
	if useCliques {
		nodeGroups = topo.BronKerbosch(g)
	} else {
		nodeGroups = topo.ConnectedComponents(g)
	}

	sortComponents(nodeGroups)

	for _, group := range nodeGroups {
		tagSet := map[subsetutils.Tag]bool{}
		for _, node := range group {
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
	confs := []any{}
	var relevant []PolicyItemWithMeta
	for i := range items {
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
				Subset: ss,
				Conf:   mergedRule,
				Origin: util_slices.Map(common.Origins(relevant, false), getMeta),
				OriginByMatches: util_maps.MapValues(common.OriginByMatches(relevant), func(_ common_api.MatchesHash, v common.Origin) core_model.ResourceMeta {
					return getMeta(v)
				}),
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
	case common_api.LegacyMeshSubsetKind():
		ss := subsetutils.Subset{}
		for k, v := range pointer.Deref(tr.Tags) {
			ss = append(ss, subsetutils.Tag{Key: k, Value: v})
		}
		return ss, nil
	case common_api.MeshService:
		return subsetutils.Subset{{Key: mesh_proto.ServiceTag, Value: serviceTagValue(tr)}}, nil
	case common_api.LegacyMeshServiceSubsetKind():
		ss := subsetutils.Subset{{Key: mesh_proto.ServiceTag, Value: serviceTagValue(tr)}}
		for k, v := range pointer.Deref(tr.Tags) {
			ss = append(ss, subsetutils.Tag{Key: k, Value: v})
		}
		return ss, nil
	default:
		return nil, errors.Errorf("can't represent %s as tags", tr.Kind)
	}
}

// policySelectsByNamespace reports whether a policy may reference a resource
// living in the given namespace. Consumer and workload-owner policies are
// namespaced, so they can only reference resources from their own namespace;
// producer/system policies are namespace-agnostic. This mirrors the dataplane
// namespace scoping applied during matching.
func policySelectsByNamespace(policyMeta, resourceMeta core_model.ResourceMeta) bool {
	switch core_model.PolicyRole(policyMeta) {
	case mesh_proto.ConsumerPolicyRole, mesh_proto.WorkloadOwnerPolicyRole:
		ns, ok := policyMeta.GetLabels()[mesh_proto.KubeNamespaceTag]
		return ok && ns == resourceMeta.GetLabels()[mesh_proto.KubeNamespaceTag]
	default:
		return true
	}
}

// serviceTagValue returns the kuma.io/service value identifying a legacy
// MeshService targetRef. Under the labels-only contract the service name lives
// in the kuma.io/display-name label, so fall back to it when name is unset.
func serviceTagValue(tr common_api.TargetRef) string {
	return pointer.Deref(tr.Labels)[mesh_proto.DisplayName]
}
