package v1alpha1

import (
	"fmt"
	"strings"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/proto"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

const (
	NameFilterBeforeErr = "must be defined. You need to pick a filter before which this one will be added"
	NameFilterAfterErr  = "must be defined. You need to pick a filter after which this one will be added"
)

func (r *MeshProxyPatchResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("appendModifications")
	for i, modification := range conf.AppendModifications {
		path := path.Index(i)
		switch {
		case modification.Cluster != nil:
			verr.AddErrorAt(path, validateClusterMod(*modification.Cluster))
		case modification.Listener != nil:
			verr.AddErrorAt(path, validateListenerMod(*modification.Listener))
		case modification.VirtualHost != nil:
			verr.AddErrorAt(path, validateVirtualHostMod(*modification.VirtualHost))
		case modification.NetworkFilter != nil:
			verr.AddErrorAt(path, validateNetworkFilterMod(*modification.NetworkFilter))
		case modification.HTTPFilter != nil:
			verr.AddErrorAt(path, validateHTTPFilterMod(*modification.HTTPFilter))
		default:
			verr.AddViolationAt(path, "at least one modification has to be defined")
		}
	}
	if len(conf.AppendModifications) == 0 {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
	}
	return verr
}

func validateClusterMod(mod ClusterMod) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("cluster")
	switch mod.Operation {
	case ModOpAdd:
		if mod.Match != nil {
			verr.AddViolationAt(path.Field("match"), validators.MustNotBeDefined)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_cluster_v3.Cluster{}))
	case ModOpPatch:
		verr.Add(validatePatchValue(path.Field("value"), mod.Value, &envoy_cluster_v3.Cluster{}))
	case ModOpRemove:
		if mod.Value != nil {
			verr.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
		}
	default:
		verr.AddViolationAt(path.Field("operation"), availableOperationsMsg(ModOpAdd, ModOpPatch, ModOpRemove))
	}
	return verr
}

func validateListenerMod(mod ListenerMod) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("listener")
	switch mod.Operation {
	case ModOpAdd:
		if mod.Match != nil {
			verr.AddViolationAt(path.Field("match"), validators.MustNotBeDefined)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_listener_v3.Listener{}))
	case ModOpPatch:
		verr.Add(validatePatchValue(path.Field("value"), mod.Value, &envoy_listener_v3.Listener{}))
	case ModOpRemove:
		if mod.Value != nil {
			verr.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
		}
	default:
		verr.AddViolationAt(path.Field("operation"), availableOperationsMsg(ModOpAdd, ModOpPatch, ModOpRemove))
	}
	return verr
}

func validateVirtualHostMod(mod VirtualHostMod) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("virtualHost")
	switch mod.Operation {
	case ModOpAdd:
		if mod.Match != nil && mod.Match.Name != nil {
			verr.AddViolationAt(path.Field("match").Field("name"), validators.MustNotBeDefined)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_route_v3.VirtualHost{}))
	case ModOpPatch:
		verr.Add(validatePatchValue(path.Field("value"), mod.Value, &envoy_route_v3.VirtualHost{}))
	case ModOpRemove:
		if mod.Value != nil {
			verr.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
		}
	default:
		verr.AddViolationAt(path.Field("operation"), availableOperationsMsg(ModOpAdd, ModOpPatch, ModOpRemove))
	}
	return verr
}

func validateHTTPFilterMod(mod HTTPFilterMod) validators.ValidationError {
	verr := validators.ValidationError{}
	path := validators.RootedAt("httpFilter")
	switch mod.Operation {
	case ModOpAddFirst, ModOpAddLast:
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_hcm_v3.HttpFilter{}))
	case ModOpAddBefore:
		if mod.Match == nil || mod.Match.Name == nil {
			verr.AddViolationAt(path.Field("match").Field("name"), NameFilterBeforeErr)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_hcm_v3.HttpFilter{}))
	case ModOpAddAfter:
		if mod.Match == nil || mod.Match.Name == nil {
			verr.AddViolationAt(path.Field("match").Field("name"), NameFilterAfterErr)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_hcm_v3.HttpFilter{}))
	case ModOpPatch:
		if mod.Match == nil || mod.Match.Name == nil {
			verr.AddViolationAt(path.Field("match").Field("name"), validators.MustBeDefined)
		}
		verr.Add(validatePatchValue(path.Field("value"), mod.Value, &envoy_hcm_v3.HttpFilter{}))
	case ModOpRemove:
		if mod.Value != nil {
			verr.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
		}
	default:
		verr.AddViolationAt(path.Field("operation"), availableOperationsMsg(ModOpAddFirst, ModOpAddLast, ModOpAddBefore, ModOpAddAfter, ModOpPatch, ModOpRemove))
	}
	return verr
}

func validateNetworkFilterMod(mod NetworkFilterMod) validators.ValidationError {
	verr := validators.ValidationError{}
	path := validators.RootedAt("networkFilter")
	switch mod.Operation {
	case ModOpAddFirst, ModOpAddLast:
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_listener_v3.Filter{}))
	case ModOpAddBefore:
		if mod.Match == nil || mod.Match.Name == nil {
			verr.AddViolationAt(path.Field("match").Field("name"), NameFilterBeforeErr)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_listener_v3.Filter{}))
	case ModOpAddAfter:
		if mod.Match == nil || mod.Match.Name == nil {
			verr.AddViolationAt(path.Field("match").Field("name"), NameFilterAfterErr)
		}
		verr.Add(validateResourceValue(path.Field("value"), mod.Value, &envoy_listener_v3.Filter{}))
	case ModOpPatch:
		if mod.Match == nil || mod.Match.Name == nil {
			verr.AddViolationAt(path.Field("match").Field("name"), validators.MustBeDefined)
		}
		verr.Add(validatePatchValue(path.Field("value"), mod.Value, &envoy_listener_v3.Filter{}))
	case ModOpRemove:
		if mod.Value != nil {
			verr.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
		}
	default:
		verr.AddViolationAt(path.Field("operation"), availableOperationsMsg(ModOpAddFirst, ModOpAddLast, ModOpAddBefore, ModOpAddAfter, ModOpPatch, ModOpRemove))
	}
	return verr
}

func validateResourceValue(path validators.PathBuilder, value *string, res proto.Message) validators.ValidationError {
	var verr validators.ValidationError
	if value != nil {
		if err := mesh.ValidateAnyResourceYAML(*value, res); err != nil {
			verr.AddViolationAt(path, fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	} else {
		verr.AddViolationAt(path, validators.MustBeDefined)
	}
	return verr
}

func validatePatchValue(path validators.PathBuilder, value *string, res proto.Message) validators.ValidationError {
	var verr validators.ValidationError
	if value != nil {
		if err := mesh.ValidateAnyResourceYAMLPatch(*value, res); err != nil {
			verr.AddViolationAt(path, fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	} else {
		verr.AddViolationAt(path, validators.MustBeDefined)
	}
	return verr
}

func availableOperationsMsg(operations ...ModOperation) string {
	var ops []string
	for _, op := range operations {
		ops = append(ops, fmt.Sprintf("%q", op))
	}
	return fmt.Sprintf("invalid operation. Available operations: %s", strings.Join(ops, ", "))
}
