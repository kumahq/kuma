package outbound

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type ResourceRule struct {
	Resource            core_model.ResourceMeta
	ResourceSectionName string
	Conf                []interface{}
	Origin              []common.Origin

	// OriginByMatches is an auxiliary structure for MeshHTTPRoute rules. It's a mapping between the rule (identified
	// by the hash of rule's matches) and the origin of the MeshHTTPRoute policy that contributed the rule.
	OriginByMatches map[common_api.MatchesHash]common.Origin
}

type ResourceRules map[kri.Identifier]ResourceRule

func (rr ResourceRules) Compute(uri kri.Identifier, reader kri.ResourceReader) *ResourceRule {
	if rule, ok := rr[uri]; ok {
		return &rule
	}

	switch uri.ResourceType {
	case meshservice_api.MeshServiceType:
	case meshexternalservice_api.MeshExternalServiceType:
	case meshmultizoneservice_api.MeshMultiZoneServiceType:
	case meshhttproute_api.MeshHTTPRouteType:
	default:
		// For other resource types no further processing can produce a valid rule, so if nothing
		// was found above we return nil
		return nil
	}

	// If the resource has a sectionName, try again without it. For MeshService, MeshExternalService,
	// and MeshMultiZoneService this means checking rules at the service level instead of a specific
	// port. MeshHTTPRoute has no sectionName, so no special handling is needed
	if uri.SectionName != "" {
		return rr.Compute(kri.NoSectionName(uri), reader)
	}

	// If still not found, try computing rules for the Mesh
	meshID := kri.Identifier{
		ResourceType: core_mesh.MeshType,
		Name:         uri.Mesh,
	}

	if mesh := reader.Get(meshID); mesh != nil {
		return rr.Compute(kri.From(mesh), reader)
	}

	return nil
}

type ToEntry interface {
	common.BaseEntry
	GetTargetRef() common_api.TargetRef
}

// BuildRules constructs ResourceRules from the given policies and resource reader.
// It first extracts 'to' entries from the policies and then builds the rules based on these entries.
func BuildRules(policies core_model.ResourceList, reader kri.ResourceReader) (ResourceRules, error) {
	entries, err := GetEntries(policies)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get 'to' entries")
	}
	return buildRules(entries, reader)
}

func GetEntries(policies core_model.ResourceList) ([]common.WithPolicyAttributes[ToEntry], error) {
	policiesWithTo, ok := common.Cast[core_model.PolicyWithToList](policies.GetItems())
	if !ok {
		return nil, nil
	}

	entries := []common.WithPolicyAttributes[ToEntry]{}

	for i, pwt := range policiesWithTo {
		for j, item := range pwt.GetToList() {
			entries = append(entries, common.WithPolicyAttributes[ToEntry]{
				Entry:     item,
				Meta:      policies.GetItems()[i].GetMeta(),
				TopLevel:  pwt.GetTargetRef(),
				RuleIndex: j,
			})
		}
	}
	return entries, nil
}

func buildRules[T interface {
	common.PolicyAttributes
	common.Entry[ToEntry]
}](list []T, reader kri.ResourceReader) (ResourceRules, error) {
	rules := ResourceRules{}

	Sort(list)

	var resolvedItems []*withResolvedResource[T]
	for _, item := range list {
		rs := resolve.TargetRef(item.GetEntry().GetTargetRef(), item.GetResourceMeta(), reader)
		for _, r := range rs {
			resolvedItems = append(resolvedItems, &withResolvedResource[T]{entry: item, rs: r})
		}
	}

	indexed := map[kri.Identifier]core_model.Resource{}
	for _, i := range resolvedItems {
		indexed[i.rs.Identifier()] = i.rs.Resource
	}

	// we could've built ResourceRule for all resources in the cluster, but we only need to build rules for resources
	// that are part of the policy to reduce the size of the ResourceRules
	for uri, resource := range indexed {
		// take only policy items that have isRelevant conf for the resource
		var relevant []T
		for _, policyItem := range resolvedItems {
			if policyItem.isRelevant(resource, uri.SectionName) {
				relevant = append(relevant, policyItem.entry)
			}
		}

		if len(relevant) > 0 {
			// merge all relevant confs into one, the order of merging is guaranteed by SortToEntries
			merged, err := merge.Entries(relevant)
			if err != nil {
				return nil, err
			}
			rules[uri] = ResourceRule{
				Resource:            resource.GetMeta(),
				ResourceSectionName: uri.SectionName,
				Conf:                merged,
				Origin:              common.Origins(relevant, true),
				OriginByMatches:     common.OriginByMatches(relevant),
			}
		}
	}

	return rules, nil
}

type withResolvedResource[T any] struct {
	entry T
	rs    *resolve.ResourceSection
}

// isRelevant returns true if the rs.resource is relevant to the other resource or section of the resource
func (rw *withResolvedResource[T]) isRelevant(other core_model.Resource, sectionName string) bool {
	switch itemDescriptorName := rw.rs.Resource.Descriptor().Name; itemDescriptorName {
	case core_mesh.MeshType:
		switch other.Descriptor().Name {
		case core_mesh.MeshType:
			return rw.rs.Resource.GetMeta().GetName() == other.GetMeta().GetName()
		default:
			return rw.rs.Resource.GetMeta().GetName() == other.GetMeta().GetMesh()
		}
	case meshservice_api.MeshServiceType, meshmultizoneservice_api.MeshMultiZoneServiceType:
		if other.Descriptor().Name != itemDescriptorName {
			return false
		}
		switch {
		case kri.WithSectionName(kri.From(rw.rs.Resource), rw.rs.SectionName) == kri.WithSectionName(kri.From(other), sectionName):
			return true
		case kri.From(rw.rs.Resource) == kri.From(other) && rw.rs.SectionName == "" && sectionName != "":
			return true
		default:
			return false
		}
	case meshexternalservice_api.MeshExternalServiceType:
		if other.Descriptor().Name != itemDescriptorName {
			return false
		}
		return kri.From(rw.rs.Resource) == kri.From(other)
	case meshhttproute_api.MeshHTTPRouteType:
		if other.Descriptor().Name != itemDescriptorName {
			return false
		}
		return kri.From(rw.rs.Resource) == kri.From(other)
	default:
		return false
	}
}
