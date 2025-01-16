package outbound

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type ResourceRule struct {
	Resource            core_model.ResourceMeta
	ResourceSectionName string
	Conf                []interface{}
	Origin              []common.Origin

	// BackendRefOriginIndex is a mapping from the rule to the origin of the BackendRefs in the rule.
	// Some policies have BackendRefs in their confs, and it's important to know what was the original policy
	// that contributed the BackendRefs to the final conf. Rule (key) is represented as a hash from rule.Matches.
	// Origin (value) is represented as an index in the Origin list. If policy doesn't have rules (i.e. MeshTCPRoute)
	// then key is an empty string "".
	BackendRefOriginIndex common.BackendRefOriginIndex
}

func (r *ResourceRule) GetBackendRefOrigin(hash common_api.MatchesHash) (core_model.ResourceMeta, bool) {
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
	return r.Origin[index].Resource, true
}

type ResourceRules map[core_model.TypedResourceIdentifier]ResourceRule

func (rr ResourceRules) Compute(uri core_model.TypedResourceIdentifier, reader common.ResourceReader) *ResourceRule {
	if rule, ok := rr[uri]; ok {
		return &rule
	}

	switch uri.ResourceType {
	case meshservice_api.MeshServiceType, meshmultizoneservice_api.MeshMultiZoneServiceType:
		// find MeshService without the sectionName and compute rules for it
		if uri.SectionName != "" {
			uriWithoutSection := uri
			uriWithoutSection.SectionName = ""
			return rr.Compute(uriWithoutSection, reader)
		}
		// find MeshService's Mesh and compute rules for it
		if mesh := reader.Get(core_mesh.MeshType, core_model.ResourceIdentifier{Name: uri.Mesh}); mesh != nil {
			return rr.Compute(UniqueKey(mesh, ""), reader)
		}
	case meshexternalservice_api.MeshExternalServiceType:
		// find MeshExternalService's Mesh and compute rules for it
		if mesh := reader.Get(core_mesh.MeshType, core_model.ResourceIdentifier{Name: uri.Mesh}); mesh != nil {
			return rr.Compute(UniqueKey(mesh, ""), reader)
		}
	case meshhttproute_api.MeshHTTPRouteType:
		// todo(lobkovilya): handle MeshHTTPRoute
	}

	return nil
}

type ToEntry interface {
	common.BaseEntry
	GetTargetRef() common_api.TargetRef
}

// BuildRules constructs ResourceRules from the given policies and resource reader.
// It first extracts 'to' entries from the policies and then builds the rules based on these entries.
func BuildRules(policies []core_model.Resource, reader common.ResourceReader) (ResourceRules, error) {
	entries, err := GetEntries(policies)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get 'to' entries")
	}
	return buildRules(entries, reader)
}

func GetEntries(policies []core_model.Resource) ([]common.WithPolicyAttributes[ToEntry], error) {
	policiesWithTo, ok := common.Cast[core_model.PolicyWithToList](policies)
	if !ok {
		return nil, nil
	}

	entries := []common.WithPolicyAttributes[ToEntry]{}

	for i, pwt := range policiesWithTo {
		for j, item := range pwt.GetToList() {
			entries = append(entries, common.WithPolicyAttributes[ToEntry]{
				Entry:     item,
				Meta:      policies[i].GetMeta(),
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
}](list []T, reader common.ResourceReader) (ResourceRules, error) {
	rules := ResourceRules{}

	Sort(list)

	var resolvedItems []*withResolvedResource[T]
	for _, item := range list {
		resolvedItems = append(resolvedItems, resolveTargetRef(item, reader)...)
	}

	indexed := map[core_model.TypedResourceIdentifier]core_model.Resource{}
	for _, i := range resolvedItems {
		indexed[UniqueKey(i.resource, i.sectionName())] = i.resource
	}

	// we could've built ResourceRule for all resources in the cluster, but we only need to build rules for resources
	// that are part of the policy to reduce the size of the ResourceRules
	for uri, resource := range indexed {
		// take only policy items that have isRelevant conf for the resource
		var relevant []T
		for _, policyItem := range resolvedItems {
			if policyItem.isRelevant(resource, uri.SectionName) {
				relevant = append(relevant, policyItem.item)
			}
		}

		if len(relevant) > 0 {
			// merge all relevant confs into one, the order of merging is guaranteed by SortToEntries
			merged, err := merge.Entries(relevant)
			if err != nil {
				return nil, err
			}
			ruleOrigins, originIndex := common.Origins(relevant, true)
			rules[uri] = ResourceRule{
				Resource:              resource.GetMeta(),
				ResourceSectionName:   uri.SectionName,
				Conf:                  merged,
				Origin:                ruleOrigins,
				BackendRefOriginIndex: originIndex,
			}
		}
	}

	return rules, nil
}

func UniqueKey(r core_model.Resource, sectionName string) core_model.TypedResourceIdentifier {
	return core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.NewResourceIdentifier(r),
		ResourceType:       r.Descriptor().Name,
		SectionName:        sectionName,
	}
}

type withResolvedResource[T interface {
	common.PolicyAttributes
	common.Entry[ToEntry]
}] struct {
	item T

	resource            core_model.Resource
	implicitSectionName string
}

func (rw *withResolvedResource[T]) sectionName() string {
	if refSectionName := rw.item.GetEntry().GetTargetRef().SectionName; refSectionName != "" {
		return refSectionName
	}
	return rw.implicitSectionName
}

// isRelevant returns true if the resolvedWrapper's resource is relevant to the other resource or section of the resource
func (rw *withResolvedResource[T]) isRelevant(other core_model.Resource, sectionName string) bool {
	switch itemDescriptorName := rw.resource.Descriptor().Name; itemDescriptorName {
	case core_mesh.MeshType:
		switch other.Descriptor().Name {
		case core_mesh.MeshType:
			return rw.resource.GetMeta().GetName() == other.GetMeta().GetName()
		default:
			return rw.resource.GetMeta().GetName() == other.GetMeta().GetMesh()
		}
	case meshservice_api.MeshServiceType, meshmultizoneservice_api.MeshMultiZoneServiceType:
		if other.Descriptor().Name != itemDescriptorName {
			return false
		}
		switch {
		case UniqueKey(rw.resource, rw.sectionName()) == UniqueKey(other, sectionName):
			return true
		case UniqueKey(rw.resource, "") == UniqueKey(other, "") && rw.sectionName() == "" && sectionName != "":
			return true
		default:
			return false
		}
	case meshexternalservice_api.MeshExternalServiceType:
		if other.Descriptor().Name != itemDescriptorName {
			return false
		}
		return UniqueKey(rw.resource, "") == UniqueKey(other, "")
	default:
		return false
	}
}

func resolveTargetRef[T interface {
	common.PolicyAttributes
	common.Entry[ToEntry]
}](item T, reader common.ResourceReader) []*withResolvedResource[T] {
	if !item.GetEntry().GetTargetRef().Kind.IsRealResource() {
		return nil
	}
	rtype := core_model.ResourceType(item.GetEntry().GetTargetRef().Kind)
	list := reader.ListOrEmpty(rtype).GetItems()

	var implicitPort uint32
	implicitLabels := map[string]string{}
	if item.GetEntry().GetTargetRef().Kind == common_api.MeshService && item.GetEntry().GetTargetRef().SectionName == "" {
		if name, namespace, port, err := parseService(item.GetEntry().GetTargetRef().Name); err == nil {
			implicitLabels[mesh_proto.KubeNamespaceTag] = namespace
			implicitLabels[mesh_proto.DisplayName] = name
			implicitPort = port
		}
	}

	labels := item.GetEntry().GetTargetRef().Labels
	if len(implicitLabels) > 0 {
		labels = implicitLabels
	}

	if len(labels) > 0 {
		var rv []*withResolvedResource[T]
		trLabels := subsetutils.NewSubset(labels)
		for _, r := range list {
			rLabels := subsetutils.NewSubset(r.GetMeta().GetLabels())
			var implicitSectionName string
			if ms, ok := r.(*meshservice_api.MeshServiceResource); ok && implicitPort != 0 {
				for _, port := range ms.Spec.Ports {
					if port.Port == implicitPort {
						implicitSectionName = port.Name
					}
				}
			}
			if trLabels.IsSubset(rLabels) {
				rv = append(rv, &withResolvedResource[T]{resource: r, item: item, implicitSectionName: implicitSectionName})
			}
		}
		return rv
	}

	ri := core_model.TargetRefToResourceIdentifier(item.GetResourceMeta(), item.GetEntry().GetTargetRef())
	if resource := reader.Get(rtype, ri); resource != nil {
		return []*withResolvedResource[T]{{resource: resource, item: item}}
	}

	return nil
}

func parseService(host string) (string, string, uint32, error) {
	// split host into <name>_<namespace>_svc_<port>
	segments := strings.Split(host, "_")

	var port uint32
	switch len(segments) {
	case 4:
		p, err := strconv.ParseInt(segments[3], 10, 32)
		if err != nil {
			return "", "", 0, err
		}
		port = uint32(p)
	case 3:
		// service less service names have no port, so we just put the reserved
		// one here to note that this service is actually
		port = mesh_proto.TCPPortReserved
	default:
		return "", "", 0, errors.Errorf("service tag in unexpected format")
	}

	name, namespace := segments[0], segments[1]
	return name, namespace, port, nil
}
