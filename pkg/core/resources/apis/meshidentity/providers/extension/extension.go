package extension

import (
	"context"
	"fmt"
	"sort"

	"github.com/kumahq/kuma/v2/pkg/core"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
)

var log = core.Log.WithName("identity-provider").WithName("extension")

// Dispatcher routes MeshIdentity Extension requests to handlers registered by
// Extension.Name. It implements providers.IdentityProvider and is registered as
// the single "Extension" provider in the identity manager.
type Dispatcher struct {
	handlers map[string]providers.IdentityProvider
}

var _ providers.IdentityProvider = &Dispatcher{}

// NewDispatcher creates a dispatcher that routes to extension sub-handlers by name.
func NewDispatcher(handlers map[string]providers.IdentityProvider) *Dispatcher {
	return &Dispatcher{handlers: handlers}
}

func (d *Dispatcher) Validate(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	handler, err := d.handler(identity)
	if err != nil {
		return err
	}
	return handler.Validate(ctx, identity)
}

func (d *Dispatcher) Initialize(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	handler, err := d.handler(identity)
	if err != nil {
		return err
	}
	return handler.Initialize(ctx, identity)
}

func (d *Dispatcher) CreateIdentity(ctx context.Context, identity *meshidentity_api.MeshIdentityResource, proxy *xds.Proxy) (*xds.WorkloadIdentity, error) {
	handler, err := d.handler(identity)
	if err != nil {
		return nil, err
	}
	return handler.CreateIdentity(ctx, identity, proxy)
}

func (d *Dispatcher) GetMeshTrustCA(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	handler, err := d.handler(identity)
	if err != nil {
		return nil, err
	}
	return handler.GetMeshTrustCA(ctx, identity)
}

func (d *Dispatcher) handler(identity *meshidentity_api.MeshIdentityResource) (providers.IdentityProvider, error) {
	ext := identity.Spec.Provider.Extension
	if ext == nil {
		return nil, fmt.Errorf("MeshIdentity %q has provider type Extension but extension config is nil", identity.GetMeta().GetName())
	}
	handler, found := d.handlers[ext.Name]
	if !found {
		return nil, fmt.Errorf("unknown extension %q, registered: %v", ext.Name, d.registeredNames())
	}
	return handler, nil
}

func (d *Dispatcher) registeredNames() []string {
	names := make([]string, 0, len(d.handlers))
	for n := range d.handlers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// LogRegistered logs the list of registered extension names at Info level.
func (d *Dispatcher) LogRegistered() {
	log.Info("extension dispatcher initialized", "registeredNames", d.registeredNames())
}
