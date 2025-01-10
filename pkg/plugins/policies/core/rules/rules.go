package rules

import (
	"encoding"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
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
	Rules map[InboundListener]Rules
}

type ToRules struct {
	Rules Rules
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

func InboundListenerHostnameFromGatewayListener(
	l *mesh_proto.MeshGateway_Listener,
	address string,
) InboundListenerHostname {
	return NewInboundListenerHostname(
		address,
		l.GetPort(),
		l.GetNonEmptyHostname(),
	)
}

type GatewayToRules struct {
	ByListenerAndHostname map[InboundListenerHostname]Rules
	ByListener            map[InboundListener]Rules
}

type GatewayRules struct {
	ToRules   GatewayToRules
	FromRules map[InboundListener]Rules
}

type SingleItemRules struct {
	Rules Rules
}

type PolicyItemWithMeta struct {
	core_model.PolicyItem
	core_model.ResourceMeta
}

// Tag is a key-value pair. If Not is true then Key != Value
type Tag struct {
	Key   string
	Value string
	Not   bool
}

// Subset represents a group of proxies
type Subset []Tag

// ContainsElement returns true if there exists a key in 'other' that matches the current set,
// and the corresponding k-v pair must match the set rule. Also, the left set rules of the current set can't make an impact.
// Empty set is a superset for all elements.
//
// For example if you have a Subset with Tags: [{key: zone, value: east, not: true}, {key: service, value: frontend, not: false}]
// an Element with k-v pairs: 1) service: frontend  2) version: zone1
// there's a k-v pair 'service: frontend' in Element that matches the set rule {key: service, value: frontend, not: false}
// the left set rule of Subset {key: zone, value: east, not: true} won't make an impact because of 'not: true'
func (ss Subset) ContainsElement(other Element) bool {
	// 1. find the overlaps of element and current subset
	// 2. verify the overlaps
	// 3. verify the left of current subset
	// 4. if no overlaps, verify if all the Subset rules are negative

	if len(ss) == 0 {
		return true
	}
	if len(other) == 0 {
		return false
	}

	hasOverlapKey := false
	for _, tag := range ss {
		otherVal, ok := other[tag.Key]
		if ok {
			hasOverlapKey = true

			// contradict
			if tag.Value == otherVal && tag.Not {
				return false
			}
			// intersect
			if tag.Value == otherVal && !tag.Not {
				continue
			}
			// intersect
			if tag.Value != otherVal && tag.Not {
				continue
			}
			// contradict
			if tag.Value != otherVal && !tag.Not {
				return false
			}
		} else if !tag.Not {
			// For those items that don't exist in element should not make an impact.
			// For example, the DP with tag {"service: frontend"} doesn't match
			// the policy with matching tags [{"service: frontend"}, {"zone": "east"}]
			return false
		}
	}

	// if the current Subset owns all of negative rules and no overlapped keys in Element,
	// we can also regard the Subset contains Element
	if !hasOverlapKey && ss.NumPositive() == 0 {
		return true
	}

	return hasOverlapKey
}

// IsSubset returns true if 'other' is a subset of the current set.
// Empty set is a superset for all subsets.
func (ss Subset) IsSubset(other Subset) bool {
	if len(ss) == 0 {
		return true
	}
	otherByKeys := map[string][]Tag{}
	for _, t := range other {
		otherByKeys[t.Key] = append(otherByKeys[t.Key], t)
	}
	for _, tag := range ss {
		oTags, ok := otherByKeys[tag.Key]
		if !ok {
			return false
		}
		for _, otherTag := range oTags {
			if !isSubset(tag, otherTag) {
				return false
			}
		}
	}
	return true
}

func isSubset(t1, t2 Tag) bool {
	switch {
	// t2={y: b} can't be a subset of t1={x: a} because point {y: b, x: c} belongs to t2, but doesn't belong to t1
	case t1.Key != t2.Key:
		return false

	// t2={y: !a} is a subset of t1={y: !b} if and only if a == b
	case t1.Not == t2.Not:
		return t1.Value == t2.Value

	// t2={y: a} is a subset of t1={y: !b} if and only if a != b
	case t1.Not:
		return t1.Value != t2.Value

	// t2={y: !a} can't be a subset of t1={y: b} because point {y: c} belongs to t2, but doesn't belong to t1
	case t2.Not:
		return false

	default:
		panic("impossible")
	}
}

// Intersect returns true if there exists an element that belongs both to 'other' and current set.
// Empty set intersects with all sets.
//
// We're using this function to check if 2 'from' rules of MeshTrafficPermission can be applied to the same client DPP.
// For example:
//
//	from:
//	  - targetRef:
//	      kind: MeshSubset
//	      tags:
//	        team: team-a
//	  - targetRef:
//	      kind: MeshSubset
//	      tags:
//	        zone: east
//
// there is a DPP with tags 'team: team-a' and 'zone: east' that's subjected to both these rules.
// So 'from[0]' and 'from[1]' have an intersection.
// However, in another example:
//
//	from:
//	 - targetRef:
//	     kind: MeshSubset
//	     tags:
//	       team: team-a
//	 - targetRef:
//	     kind: MeshSubset
//	     tags:
//	       team: team-b
//	       zone: east
//
// there is no DPP that'd hit both 'from[0]' and 'from[1]'. So in this case they don't have an intersection.
func (ss Subset) Intersect(other Subset) bool {
	if len(ss) == 0 || len(other) == 0 {
		return true
	}
	otherByKeysOnlyPositive := map[string][]Tag{}
	for _, t := range other {
		if t.Not {
			continue
		}
		otherByKeysOnlyPositive[t.Key] = append(otherByKeysOnlyPositive[t.Key], t)
	}
	for _, tag := range ss {
		if tag.Not {
			continue
		}
		oTags, ok := otherByKeysOnlyPositive[tag.Key]
		if !ok {
			continue
		}
		for _, otherTag := range oTags {
			if otherTag != tag {
				return false
			}
		}
	}
	return true
}

func (ss Subset) WithTag(key, value string, not bool) Subset {
	return append(ss, Tag{Key: key, Value: value, Not: not})
}

func MeshSubset() Subset {
	return Subset{}
}

func MeshService(name string) Subset {
	return Subset{{
		Key: mesh_proto.ServiceTag, Value: name,
	}}
}

func SubsetFromTags(tags map[string]string) Subset {
	subset := Subset{}
	for k, v := range tags {
		subset = append(subset, Tag{Key: k, Value: v})
	}
	return subset
}

type Element map[string]string

func (e Element) WithKeyValue(key, value string) Element {
	c := maps.Clone(e)
	if c == nil {
		c = Element{}
	}

	c[key] = value
	return c
}

func MeshElement() Element {
	return Element{}
}

func MeshServiceElement(name string) Element {
	return Element{mesh_proto.ServiceTag: name}
}

// NumPositive returns a number of tags without negation
func (ss Subset) NumPositive() int {
	pos := 0
	for _, t := range ss {
		if !t.Not {
			pos++
		}
	}
	return pos
}

func (ss Subset) IndexOfPositive() int {
	for i, t := range ss {
		if !t.Not {
			return i
		}
	}
	return -1
}

// Rule contains a configuration for the given Subset. When rule is an inbound rule (from),
// then Subset represents a group of clients. When rule is an outbound (to) then Subset
// represents destinations.
type Rule struct {
	Subset Subset
	Conf   interface{}
	Origin []core_model.ResourceMeta
}

type Rules []*Rule

// Compute returns Rule for the given element.
func (rs Rules) Compute(element Element) *Rule {
	for _, rule := range rs {
		if rule.Subset.ContainsElement(element) {
			return rule
		}
	}
	return nil
}

// ComputeConf returns configuration for the given element.
func ComputeConf[T any](rs Rules, element Element) *T {
	computed := rs.Compute(element)
	if computed != nil {
		return pointer.To(computed.Conf.(T))
	}

	return nil
}

func BuildFromRules(
	matchedPoliciesByInbound map[InboundListener][]core_model.Resource,
) (FromRules, error) {
	rulesByInbound := map[InboundListener]Rules{}
	for inbound, policies := range matchedPoliciesByInbound {
		fromList := []PolicyItemWithMeta{}
		for _, p := range policies {
			policyWithFrom, ok := p.GetSpec().(core_model.PolicyWithFromList)
			if !ok {
				return FromRules{}, nil
			}
			fromList = append(fromList, BuildPolicyItemsWithMeta(policyWithFrom.GetFromList(), p.GetMeta())...)
		}
		rules, err := BuildRules(fromList)
		if err != nil {
			return FromRules{}, err
		}
		rulesByInbound[inbound] = rules
	}
	return FromRules{
		Rules: rulesByInbound,
	}, nil
}

func BuildToRules(matchedPolicies []core_model.Resource, httpRoutes []core_model.Resource) (ToRules, error) {
	toList := []PolicyItemWithMeta{}
	for _, mp := range matchedPolicies {
		tl, err := buildToList(mp, httpRoutes)
		if err != nil {
			return ToRules{}, err
		}
		toList = append(toList, BuildPolicyItemsWithMeta(tl, mp.GetMeta())...)
	}

	rules, err := BuildRules(toList)
	if err != nil {
		return ToRules{}, err
	}

	return ToRules{Rules: rules}, nil
}

func BuildGatewayRules(
	matchedPoliciesByInbound map[InboundListener][]core_model.Resource,
	matchedPoliciesByListener map[InboundListenerHostname][]core_model.Resource,
	httpRoutes []core_model.Resource,
) (GatewayRules, error) {
	toRulesByInbound := map[InboundListener]Rules{}
	toRulesByListenerHostname := map[InboundListenerHostname]Rules{}
	for listener, policies := range matchedPoliciesByListener {
		toRules, err := BuildToRules(policies, httpRoutes)
		if err != nil {
			return GatewayRules{}, err
		}
		toRulesByListenerHostname[listener] = toRules.Rules
	}
	for inbound, policies := range matchedPoliciesByInbound {
		toRules, err := BuildToRules(policies, httpRoutes)
		if err != nil {
			return GatewayRules{}, err
		}
		toRulesByInbound[inbound] = toRules.Rules
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

func buildToList(p core_model.Resource, httpRoutes []core_model.Resource) ([]core_model.PolicyItem, error) {
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
				rv = append(rv, &artificialPolicyItem{
					targetRef: common_api.TargetRef{
						Kind: common_api.MeshServiceSubset,
						Name: mhrRules.TargetRef.Name,
						Tags: map[string]string{
							RuleMatchesHashTag: matchesHash,
						},
					},
					conf: to.GetDefault(),
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

func BuildPolicyItemsWithMeta(items []core_model.PolicyItem, meta core_model.ResourceMeta) []PolicyItemWithMeta {
	var result []PolicyItemWithMeta
	for _, item := range items {
		result = append(result, PolicyItemWithMeta{
			PolicyItem:   item,
			ResourceMeta: meta,
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

	// 1. Convert list of rules into the list of subsets
	var subsets []Subset
	for _, item := range list {
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
		tagSet := map[Tag]bool{}
		for _, node := range nodes {
			for _, t := range subsets[node.ID()] {
				tagSet[t] = true
			}
		}

		tags := []Tag{}
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
		iter := NewSubsetIter(tags)
		for {
			ss := iter.Next()
			if ss == nil {
				break
			}
			// 5. For each combination determine a configuration
			confs := []interface{}{}
			distinctOrigins := map[core_model.ResourceKey]core_model.ResourceMeta{}
			for i := 0; i < len(list); i++ {
				item := list[i]
				itemSubset, err := asSubset(item.GetTargetRef())
				if err != nil {
					return nil, err
				}
				if itemSubset.IsSubset(ss) {
					confs = append(confs, item.GetDefault())
					distinctOrigins[core_model.MetaToResourceKey(item.ResourceMeta)] = item.ResourceMeta
				}
			}
			merged, err := MergeConfs(confs)
			if err != nil {
				return nil, err
			}
			if merged != nil {
				origins := maps.Values(distinctOrigins)
				sort.Slice(origins, func(i, j int) bool {
					return origins[i].GetName() < origins[j].GetName()
				})
				for _, mergedRule := range merged {
					rules = append(rules, &Rule{
						Subset: ss,
						Conf:   mergedRule,
						Origin: origins,
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

func asSubset(tr common_api.TargetRef) (Subset, error) {
	switch tr.Kind {
	case common_api.Mesh:
		return Subset{}, nil
	case common_api.MeshSubset:
		ss := Subset{}
		for k, v := range tr.Tags {
			ss = append(ss, Tag{Key: k, Value: v})
		}
		return ss, nil
	case common_api.MeshService:
		return Subset{{Key: mesh_proto.ServiceTag, Value: tr.Name}}, nil
	case common_api.MeshServiceSubset:
		ss := Subset{{Key: mesh_proto.ServiceTag, Value: tr.Name}}
		for k, v := range tr.Tags {
			ss = append(ss, Tag{Key: k, Value: v})
		}
		return ss, nil
	default:
		return nil, errors.Errorf("can't represent %s as tags", tr.Kind)
	}
}

type SubsetIter struct {
	current  []Tag
	finished bool
}

func NewSubsetIter(tags []Tag) *SubsetIter {
	return &SubsetIter{
		current: tags,
	}
}

// Next returns the next subset of the partition. When reaches the end Next returns 'nil'
func (c *SubsetIter) Next() Subset {
	if c.finished {
		return nil
	}
	for {
		hasNext := c.next()
		if !hasNext {
			c.finished = true
			return c.simplified()
		}
		if result := c.simplified(); result != nil {
			return result
		}
	}
}

func (c *SubsetIter) next() bool {
	for idx := 0; idx < len(c.current); idx++ {
		if c.current[idx].Not {
			c.current[idx].Not = false
		} else {
			c.current[idx].Not = true
			return true
		}
	}
	return false
}

// simplified returns copy of c.current and deletes redundant tags, for example:
//   - env: dev
//   - env: !prod
//
// could be simplified to:
//   - env: dev
//
// If tags are contradicted (same keys have different positive value) then the function
// returns nil.
func (c *SubsetIter) simplified() Subset {
	result := Subset{}

	ssByKey := map[string]Subset{}
	keyOrder := []string{}
	for _, t := range c.current {
		if _, ok := ssByKey[t.Key]; !ok {
			keyOrder = append(keyOrder, t.Key)
		}
		ssByKey[t.Key] = append(ssByKey[t.Key], Tag{Key: t.Key, Value: t.Value, Not: t.Not})
	}

	for _, key := range keyOrder {
		ss := ssByKey[key]
		positive := ss.NumPositive()
		switch {
		case positive == 0:
			result = append(result, ss...)
		case positive == 1:
			result = append(result, ss[ss.IndexOfPositive()])
		case positive >= 2:
			// contradicted, at least 2 positive values for the same key, i.e 'key1: value1' and 'key1: value2'
			return nil
		}
	}

	return result
}
