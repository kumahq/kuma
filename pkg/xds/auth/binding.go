package auth

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

// Scope says who chooses the labels a credential doesn't prove.
type Scope int

const (
	// ScopeNone: anyone holding the credential, so only proven labels are trusted.
	ScopeNone Scope = iota
	// ScopeNamespace: whoever can write workloads in the proven namespace.
	ScopeNamespace
)

// Binding is what an authenticated credential vouches for about a proxy's
// labels. A workload identity may only be built from vouched-for labels,
// otherwise a proxy could claim another workload's identity by declaring its label.
// Labels whose value the control plane chooses are always trusted.
type Binding struct {
	// Unauthenticated means authentication is disabled and nothing was verified.
	Unauthenticated bool
	// Labels are the labels the credential proves, with the values it allows.
	Labels mesh_proto.MultiValueTagSet
	Scope  Scope
}

// Other control-plane-owned labels, like kuma.io/display-name or the zone proxy
// listener labels, are computed from what the proxy sends, so they need a binding.
var setByControlPlaneState = func() map[string]struct{} {
	keys := map[string]struct{}{}
	for _, key := range labels.SetByControlPlaneState() {
		keys[key] = struct{}{}
	}
	return keys
}()

// trusts reports whether the binding vouches for the label having this value.
func (b Binding) trusts(key, value string) bool {
	if b.Unauthenticated {
		return true
	}
	if _, ok := setByControlPlaneState[key]; ok {
		return true
	}
	if allowed, ok := b.Labels[key]; ok {
		return allowed[value]
	}
	// Namespace writers choose the other labels, so they should only count when
	// the SPIFFE ID also carries the namespace. That is not enforced yet.
	return b.Scope == ScopeNamespace
}

// ValidateIdentityLabels fails unless the binding vouches for every label the
// SPIFFE ID is rendered from (read) that the proxy declares.
func ValidateIdentityLabels(b Binding, dpLabels map[string]string, read []string) error {
	for _, key := range read {
		value, ok := dpLabels[key]
		if !ok || b.trusts(key, value) {
			continue
		}
		switch key {
		case metadata.KumaWorkload:
			return errors.Errorf("SPIFFE ID of the dataplane is built from its %q label %q, which the token does not bind, so it requires a workload-bound token", key, value)
		case mesh_proto.DisplayName:
			return errors.Errorf("SPIFFE ID of the dataplane is built from its %q label %q, which the token does not bind, so it requires a name-bound token", key, value)
		}
		return errors.Errorf("SPIFFE ID of the dataplane is built from its %q label %q, which the token does not bind, so it requires a token with that tag", key, value)
	}
	return nil
}

type bindingCtxKey struct{}

func ContextWithBinding(ctx context.Context, b Binding) context.Context {
	return context.WithValue(ctx, bindingCtxKey{}, b)
}

func BindingFromContext(ctx context.Context) (Binding, bool) {
	if ctx == nil {
		return Binding{}, false
	}
	b, ok := ctx.Value(bindingCtxKey{}).(Binding)
	return b, ok
}
