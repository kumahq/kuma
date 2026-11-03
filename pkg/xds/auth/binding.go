package auth

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

// Binding is what an authenticated credential vouches for about a proxy's
// labels. A workload identity may only be built from vouched-for labels,
// otherwise a proxy could claim another workload's identity by declaring its label.
type Binding struct {
	// AllLabels vouches for every label: the platform attests them (Kubernetes)
	// or authentication is disabled.
	AllLabels bool
	// Workload is the kuma.io/workload value the credential is bound to.
	Workload string
	// Tags are the values the credential allows, per label.
	Tags mesh_proto.MultiValueTagSet
}

// Allows reports whether the binding vouches for the label having this value.
func (b Binding) Allows(key, value string) bool {
	if b.AllLabels {
		return true
	}
	bound := false
	if key == metadata.KumaWorkload && b.Workload != "" {
		if value != b.Workload {
			return false
		}
		bound = true
	}
	if allowed, ok := b.Tags[key]; ok {
		if !allowed[value] {
			return false
		}
		bound = true
	}
	return bound
}

// ValidateIdentityLabels fails unless the binding vouches for every label the
// SPIFFE ID is rendered from (read) that the proxy declares.
func ValidateIdentityLabels(b Binding, labels map[string]string, read []string) error {
	for _, key := range read {
		value, ok := labels[key]
		if !ok || b.Allows(key, value) {
			continue
		}
		if key == metadata.KumaWorkload {
			return errors.Errorf("SPIFFE ID of the dataplane is built from its %q label %q, which the token does not bind, so it requires a workload-bound token", key, value)
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
