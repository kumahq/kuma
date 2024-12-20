package rules

import (
	"encoding"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
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
	Rules         Rules
	ResourceRules ResourceRules
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

type PolicyItemWithMeta struct {
	core_model.PolicyItem
	core_model.ResourceMeta
	TopLevel  common_api.TargetRef
	RuleIndex int
}

// Tag is a key-value pair. If Not is true then Key != Value
type Tag struct {
	Key   string
	Value string
	Not   bool
}

// Subset represents a group of proxies
type Subset []Tag

func NewSubset(m map[string]string) Subset {
	var s Subset
	for _, k := range util_maps.SortedKeys(m) {
		s = append(s, Tag{Key: k, Value: m[k]})
	}
	return s
}

func (ss Subset) ContainsElement(element Element) bool {
	// 1. find the overlaps of element and current subset
	// 2. verify the overlaps
	// 3. verify the left of current subset

	if len(ss) == 0 {
		return true
	}

	overlapKeyCount := 0
	for _, tag := range ss {
		tmpVal, ok := element[tag.Key]
		if ok {
			overlapKeyCount++

			// contradict
			if tag.Value == tmpVal && tag.Not {
				return false
			}
			// intersect
			if tag.Value == tmpVal && !tag.Not {
				continue
			}
			// intersect
			if tag.Value != tmpVal && tag.Not {
				continue
			}
			// contradict
			if tag.Value != tmpVal && !tag.Not {
				return false
			}
		} else {
			// for those items that don't exist in element should not make an impact
			if !tag.Not {
				return false
			}
		}
	}

	// no overlap means no connections
	if overlapKeyCount == 0 {
		return false
	}

	return true
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

func MeshExternalService(name string) Subset {
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
	if e == nil {
		e = Element{}
	}

	e[key] = value
	return e
}

func MeshElement() Element {
	return Element{}
}

func MeshServiceElement(name string) Element {
	return Element{mesh_proto.ServiceTag: name}
}

func MeshExternalServiceElement(name string) Element {
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

	// BackendRefOriginIndex is a mapping from the rule to the origin of the BackendRefs in the rule.
	// Some policies have BackendRefs in their confs, and it's important to know what was the original policy
	// that contributed the BackendRefs to the final conf. Rule (key) is represented as a hash from rule.Matches.
	// Origin (value) is represented as an index in the Origin list. If policy doesn't have rules (i.e. MeshTCPRoute)
	// then key is an empty string "".
	BackendRefOriginIndex BackendRefOriginIndex
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

func (rs Rules) NewCompute(element Element) *Rule {
	for _, rule := range rs {
		if rule.Subset.ContainsElement(element) {
			return rule
		}
	}
	return nil
}

func NewComputeConf[T any](rs Rules, element Element) *T {
	computed := rs.NewCompute(element)
	if computed != nil {
		return pointer.To(computed.Conf.(T))
	}

	return nil
}

// Compute returns configuration for the given subset.
func (rs Rules) Compute(sub Subset) *Rule {
	for _, rule := range rs {
		if rule.Subset.IsSubset(sub) {
			return rule
		}
	}
	return nil
}

func ComputeConf[T any](rs Rules, sub Subset) *T {
	if computed := rs.Compute(sub); computed != nil {
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
			fromList = append(fromList, BuildPolicyItemsWithMeta(policyWithFrom.GetFromList(), p.GetMeta(), policyWithFrom.GetTargetRef())...)
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

type ResourceReader interface {
	Get(resourceType core_model.ResourceType, ri core_model.ResourceIdentifier) core_model.Resource
	ListOrEmpty(resourceType core_model.ResourceType) core_model.ResourceList
}

func BuildToRules(matchedPolicies []core_model.Resource, reader ResourceReader) (ToRules, error) {
	toList, err := BuildToList(matchedPolicies, reader)
	if err != nil {
		return ToRules{}, err
	}

	rules, err := BuildRules(toList)
	if err != nil {
		return ToRules{}, err
	}

	resourceRules, err := BuildResourceRules(toList, reader)
	if err != nil {
		return ToRules{}, err
	}

	return ToRules{Rules: rules, ResourceRules: resourceRules}, nil
}

func BuildToList(matchedPolicies []core_model.Resource, reader ResourceReader) ([]PolicyItemWithMeta, error) {
	toList := []PolicyItemWithMeta{}
	for _, mp := range matchedPolicies {
		tl, err := buildToList(mp, reader.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).GetItems())
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
	matchedPoliciesByInbound map[InboundListener][]core_model.Resource,
	matchedPoliciesByListener map[InboundListenerHostname][]core_model.Resource,
	reader ResourceReader,
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
	var subsets []Subset
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
				merged, err := MergeConfs(confs)
				if err != nil {
					return nil, err
				}
				ruleOrigins, originIndex := origins(relevant, false)
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
