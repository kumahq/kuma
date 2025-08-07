package outbound

import (
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// ResourceContext represents a hierarchical resource structure and
// is always ready to return the appropriate conf by using Conf method.
// The RootContext is always the mesh. As we're iterating over resource in ResourceSet
// and going deeper to configure Listeners/Routes, we need to add more resources
// to the ResourceContext by using WithID method.
//
// For example:
// 1. At the beginning of Apply method in plugin.go rctx := RootContext()
// 2. As we start iterating over outbound listeners, rctx = rctx.WithID(outboundListenerKRI)
// 3. As we start iterating over the routes of the outbound listener, rctx = rctx.WithID(routeKRI)
//
// At any moment we can call rctx.Conf() and get right configuration.
type ResourceContext[T any] struct {
	ids      []kri.Identifier
	rules    ResourceRules
	fallback T
}

func AsResourceContext[T any](conf T) *ResourceContext[T] {
	return &ResourceContext[T]{
		rules:    ResourceRules{},
		fallback: conf,
	}
}

func RootContext[T any](mesh *core_mesh.MeshResource, rules ResourceRules) *ResourceContext[T] {
	return &ResourceContext[T]{
		ids: []kri.Identifier{
			kri.From(mesh),
		},
		rules: rules,
	}
}

// WithID creates a new ResourceContext with the provided identifier,
// giving it higher priority during configuration lookup.
// This allows for more specific configurations to override more general ones.
func (rc *ResourceContext[T]) WithID(id kri.Identifier) *ResourceContext[T] {
	newRc := &ResourceContext[T]{
		ids:   []kri.Identifier{id},
		rules: rc.rules,
	}
	newRc.ids = append(newRc.ids, rc.ids...)
	return newRc
}

// Conf retrieves the configuration of type T by searching through the identifiers
// in priority order (first to last) and returning the first matching configuration.
// If no matching configuration is found, it returns a zero value of type T.
func (rc *ResourceContext[T]) Conf() T {
	for _, id := range rc.ids {
		if rule, ok := rc.rules[id]; ok {
			return rule.Conf[0].(T)
		}
	}
	return rc.fallback
}

func (rc *ResourceContext[T]) ResourceRule() *ResourceRule {
	for _, id := range rc.ids {
		if rule, ok := rc.rules[id]; ok {
			return &rule
		}
	}
	return nil
}

func (rc *ResourceContext[T]) DirectConf() (T, bool) {
	if len(rc.ids) != 0 {
		if rule, ok := rc.rules[rc.ids[0]]; ok {
			return rule.Conf[0].(T), true
		}
	}
	var zero T
	return zero, false
}
