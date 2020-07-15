package mesh

import (
	"fmt"
	"strings"

	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/envoy"
)

var availableProfiles map[string]bool
var availableProfilesMsg string

func init() {
	profiles := []string{}
	availableProfiles = map[string]bool{}
	for _, profile := range AvailableProfiles {
		availableProfiles[profile] = true
		profiles = append(profiles, profile)
	}
	availableProfilesMsg = strings.Join(profiles, ",")
}

func (t *ProxyTemplateResource) Validate() error {
	var verr validators.ValidationError
	verr.Add(validateSelectors(t.Spec.Selectors))
	verr.AddError("conf", validateConfig(t.Spec.Conf))
	return verr.OrNil()
}

func validateConfig(conf *mesh_proto.ProxyTemplate_Conf) validators.ValidationError {
	var verr validators.ValidationError
	verr.Add(validateImports(conf.GetImports()))
	verr.Add(validateResources(conf.GetResources()))
	for i, modification := range conf.GetModifications() {
		verr.AddErrorAt(validators.RootedAt("modifications").Index(i), validateModification(modification))
	}
	return verr
}

func validateModification(modification *mesh_proto.ProxyTemplate_Modifications) validators.ValidationError {
	verr := validators.ValidationError{}
	switch modification.Type.(type) {
	case *mesh_proto.ProxyTemplate_Modifications_Cluster_:
		verr.AddError("cluster", validateClusterModification(modification.GetCluster()))
	case *mesh_proto.ProxyTemplate_Modifications_Listener_:
		verr.AddError("listener", validateListenerModification(modification.GetListener()))
	case *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_:
		verr.AddError("networkFilter", validateNetworkFilterModification(modification.GetNetworkFilter()))
	case *mesh_proto.ProxyTemplate_Modifications_HttpFilter_:
		verr.AddError("httpFilter", validateHTTPFilterModification(modification.GetHttpFilter()))
	}
	return verr
}

func validateHTTPFilterModification(filterMod *mesh_proto.ProxyTemplate_Modifications_HttpFilter) validators.ValidationError {
	verr := validators.ValidationError{}
	validateResource := func() {
		if err := ValidateResourceYAML(&envoy_hcm.HttpFilter{}, filterMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	}
	switch filterMod.Operation {
	case mesh_proto.OpAddFirst:
		validateResource()
	case mesh_proto.OpAddLast:
		validateResource()
	case mesh_proto.OpAddBefore:
		if filterMod.GetMatch().GetName() == "" {
			verr.AddViolation("match.name", "cannot be empty. You need to pick a filter before which this one will be added")
		}
		validateResource()
	case mesh_proto.OpAddAfter:
		if filterMod.GetMatch().GetName() == "" {
			verr.AddViolation("match.name", "cannot be empty. You need to pick a filter after which this one will be added")
		}
		validateResource()
	case mesh_proto.OpPatch:
		if filterMod.GetMatch().GetName() == "" {
			verr.AddViolation("match.name", "cannot be empty")
		}
		if err := ValidateResourceYAMLPatch(&envoy_hcm.HttpFilter{}, filterMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	case mesh_proto.OpRemove:
	default:
		verr.AddViolation("operation", fmt.Sprintf("invalid operation. Available operations: %q, %q, %q, %q, %q", mesh_proto.OpAddFirst, mesh_proto.OpAddLast, mesh_proto.OpAddBefore, mesh_proto.OpAddAfter, mesh_proto.OpPatch))
	}
	return verr
}

func validateListenerModification(listenerMod *mesh_proto.ProxyTemplate_Modifications_Listener) validators.ValidationError {
	verr := validators.ValidationError{}
	switch listenerMod.Operation {
	case mesh_proto.OpAdd:
		if listenerMod.Match != nil {
			verr.AddViolation("match", "cannot be defined")
		}
		if err := ValidateResourceYAML(&envoy_api.Listener{}, listenerMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	case mesh_proto.OpPatch:
		if err := ValidateResourceYAMLPatch(&envoy_api.Listener{}, listenerMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	case mesh_proto.OpRemove:
	default:
		verr.AddViolation("operation", fmt.Sprintf("invalid operation. Available operations: %q, %q, %q", mesh_proto.OpAdd, mesh_proto.OpPatch, mesh_proto.OpRemove))
	}
	return verr
}

func validateClusterModification(clusterMod *mesh_proto.ProxyTemplate_Modifications_Cluster) validators.ValidationError {
	verr := validators.ValidationError{}
	switch clusterMod.Operation {
	case mesh_proto.OpAdd:
		if clusterMod.Match != nil {
			verr.AddViolation("match", "cannot be defined")
		}
		if err := ValidateResourceYAML(&envoy_api.Cluster{}, clusterMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	case mesh_proto.OpPatch:
		if err := ValidateResourceYAMLPatch(&envoy_api.Cluster{}, clusterMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	case mesh_proto.OpRemove:
	default:
		verr.AddViolation("operation", fmt.Sprintf("invalid operation. Available operations: %q, %q, %q", mesh_proto.OpAdd, mesh_proto.OpPatch, mesh_proto.OpRemove))
	}
	return verr
}

func validateNetworkFilterModification(networkFilterMod *mesh_proto.ProxyTemplate_Modifications_NetworkFilter) validators.ValidationError {
	verr := validators.ValidationError{}
	validateResource := func() {
		if err := ValidateResourceYAML(&envoy_listener.Filter{}, networkFilterMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	}
	switch networkFilterMod.Operation {
	case mesh_proto.OpAddFirst:
		validateResource()
	case mesh_proto.OpAddLast:
		validateResource()
	case mesh_proto.OpAddBefore:
		if networkFilterMod.GetMatch().GetName() == "" {
			verr.AddViolation("match.name", "cannot be empty. You need to pick a filter before which this one will be added")
		}
		validateResource()
	case mesh_proto.OpAddAfter:
		if networkFilterMod.GetMatch().GetName() == "" {
			verr.AddViolation("match.name", "cannot be empty. You need to pick a filter after which this one will be added")
		}
		validateResource()
	case mesh_proto.OpPatch:
		if networkFilterMod.GetMatch().GetName() == "" {
			verr.AddViolation("match.name", "cannot be empty")
		}
		if err := ValidateResourceYAMLPatch(&envoy_listener.Filter{}, networkFilterMod.Value); err != nil {
			verr.AddViolation("value", fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	case mesh_proto.OpRemove:
	default:
		verr.AddViolation("operation", fmt.Sprintf("invalid operation. Available operations: %q, %q, %q, %q, %q", mesh_proto.OpAddFirst, mesh_proto.OpAddLast, mesh_proto.OpAddBefore, mesh_proto.OpAddAfter, mesh_proto.OpPatch))
	}
	return verr
}

func validateImports(imports []string) validators.ValidationError {
	var verr validators.ValidationError
	for i, imp := range imports {
		if imp == "" {
			verr.AddViolationAt(validators.RootedAt("imports").Index(i), "cannot be empty")
			continue
		}
		if !availableProfiles[imp] {
			verr.AddViolationAt(validators.RootedAt("imports").Index(i), fmt.Sprintf("profile not found. Available profiles: %s", availableProfilesMsg))
		}
	}
	return verr
}

func validateResources(resources []*mesh_proto.ProxyTemplateRawResource) validators.ValidationError {
	var verr validators.ValidationError
	for i, resource := range resources {
		if resource.Name == "" {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("name"), "cannot be empty")
		}
		if resource.Version == "" {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("version"), "cannot be empty")
		}
		if resource.Resource == "" {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("resource"), "cannot be empty")
		} else if _, err := envoy.ResourceFromYaml(resource.Resource); err != nil {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("resource"), fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	}
	return verr
}

func validateSelectors(selectors []*mesh_proto.Selector) validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("selectors"), selectors, ValidateSelectorsOpts{
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireService:       true,
			RequireAtLeastOneTag: true,
		},
	})
}
